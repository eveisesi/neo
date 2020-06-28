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

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/go-redis/redis/v7"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	app := core.New()

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

	r.Group(func(r chi.Router) {
		r.Use(NewStructuredLogger(s.logger))
		r.Use(s.Dataloaders)

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
			},
		})

		gqlhandler := handler.New(schema)
		gqlhandler.AddTransport(transport.GET{})
		gqlhandler.AddTransport(transport.POST{})
		gqlhandler.AddTransport(transport.Websocket{})
		gqlhandler.Use(extension.Introspection{})
		gqlhandler.Use(extension.AutomaticPersistedQuery{
			Cache: &GQLCache{client: s.redis, ttl: time.Hour * 24},
		})

		gqlhandler.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			opCtx := graphql.GetOperationContext(ctx)

			entry := s.logger.WithField("operationName", opCtx.Operation.Name)
			for i, v := range opCtx.Variables {
				entry = entry.WithField(fmt.Sprintf("var.%s", i), v)
			}

			entry.Println()

			return next(ctx)
		})

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				next.ServeHTTP(w, r)
			})
		})

		r.Handle("/query", gqlhandler)

	})

	r.Group(func(r chi.Router) {
		r.Use(Cors)
		r.Use(s.RateLimiter)
		r.Get("/auth/state", s.handleGetState)
		r.Post("/auth/token", s.handlePostCode)
		r.Handle("/top/metrics", promhttp.Handler())

	})

	return r

}

// GracefullyShutdown gracefully shuts down the HTTP API.
func (s *Server) GracefullyShutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) WriteSuccess(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	if status != 0 {
		w.WriteHeader(status)
	}

	return json.NewEncoder(w).Encode(data)
}

func (s *Server) WriteError(w http.ResponseWriter, code int, err error) error {
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

	return json.NewEncoder(w).Encode(res)
}
