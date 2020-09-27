package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v7"

	core "github.com/eveisesi/neo/app"
	"github.com/eveisesi/neo/graphql/resolvers"
	"github.com/eveisesi/neo/graphql/service"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/search"
	"github.com/eveisesi/neo/services/token"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Server struct {
	server *http.Server
	logger *logrus.Logger
	redis  *redis.Client

	alliance    alliance.Service
	token       token.Service
	character   character.Service
	corporation corporation.Service
	killmail    killmail.Service
	search      search.Service
	universe    universe.Service
}

func Action(c *cli.Context) {
	app := core.New("server", false)

	server := NewServer(
		app.Config.ServerPort,
		app.Logger,
		app.Redis,
		app.Alliance,
		app.Character,
		app.Corporation,
		app.Killmail,
		app.Search,
		app.Token,
		app.Universe,
	)
	app.Logger.WithField("port", app.Config.ServerPort).Info("attempting to start server...")
	go cleanUpVisitors()
	go func() {
		if err := server.server.ListenAndServe(); err != nil {
			app.Logger.WithError(err).Fatal("unable to start server")
			return
		}
	}()

	app.Logger.WithField("port", app.Config.ServerPort).Info("server started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	app.Logger.Info("attempting to gracefully shutdown server")

	err := server.GracefullyShutdown(context.Background())
	if err != nil {
		app.Logger.WithError(err).Error("unable to start serve")
	}

	app.Logger.Info("server gracefully shutdown")

}

func NewServer(
	port uint,
	logger *logrus.Logger,
	redis *redis.Client,
	alliance alliance.Service,
	character character.Service,
	corporation corporation.Service,
	killmail killmail.Service,
	search search.Service,
	token token.Service,
	universe universe.Service,
) *Server {

	visitors = make(map[string]*visitor)

	s := Server{
		logger: logger,
		redis:  redis,

		alliance:    alliance,
		character:   character,
		corporation: corporation,
		killmail:    killmail,
		search:      search,
		token:       token,
		universe:    universe,
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		Handler:      s.RegisterRoutes(),
	}

	return &s
}

func (s *Server) RegisterRoutes() *chi.Mux {

	r := chi.NewRouter()
	r.Use(Cors)

	r.Group(func(r chi.Router) {
		r.Use(s.Dataloaders)
		r.Use(NewStructuredLogger(s.logger))

		schema := service.NewExecutableSchema(service.Config{
			Resolvers: &resolvers.Resolver{
				Services: resolvers.Services{
					Killmail:    s.killmail,
					Alliance:    s.alliance,
					Corporation: s.corporation,
					Character:   s.character,
					Universe:    s.universe,
					Search:      s.search,
				},
				Dataloader: CtxLoaders,
				Logger:     s.logger,
				Redis:      s.redis,
			},
		})

		gqlhandler := handler.New(schema)
		gqlhandler.AddTransport(transport.GET{})
		gqlhandler.AddTransport(transport.POST{})
		gqlhandler.AddTransport(transport.Websocket{
			KeepAlivePingInterval: time.Second * 5,
			Upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		})
		gqlhandler.Use(extension.Introspection{})

		gqlhandler.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {

			entry, ok := ctx.Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry)
			if !ok {
				fmt.Println("entry missing")
				return next(ctx)
			}

			opCtx := graphql.GetOperationContext(ctx)
			entry.Logger = entry.Logger.WithField("requestType", opCtx.Operation.Name)

			for _, s := range opCtx.Operation.SelectionSet {
				var field *ast.Field
				var ok bool

				if field, ok = s.(*ast.Field); !ok {
					continue
				}

				for _, arg := range field.Arguments {
					value, err := arg.Value.Value(opCtx.Variables)
					if err != nil {
						continue
					}
					entry.Logger = entry.Logger.WithField(fmt.Sprintf("%s_%s", field.Name, arg.Name), value)
				}

			}

			return next(ctx)
		})

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				next.ServeHTTP(w, r)
			})
		})

		r.Handle("/query", gqlhandler)
		r.Handle("/query/playground", playground.Handler("NEO GraphQL Playground", "/query"))
		r.Get("/search", s.handleSearchRequest)

	})

	r.Group(func(r chi.Router) {
		r.Use(s.RateLimiter)
		r.Get("/auth/state", s.handleGetState)
		r.Post("/auth/token", s.handlePostCode)
	})

	return r

}

// GracefullyShutdown gracefully shuts down the HTTP API.
func (s *Server) GracefullyShutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if status != 0 {
		w.WriteHeader(status)
	}

	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) WriteError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application-type/json")
	w.WriteHeader(code)

	if err == nil {
		err = errors.New(http.StatusText(code))
	}

	res := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	_ = json.NewEncoder(w).Encode(res)
}
