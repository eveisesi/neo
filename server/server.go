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

	"github.com/99designs/gqlgen/handler"
	"github.com/go-redis/redis"

	core "github.com/eveisesi/neo/app"
	"github.com/eveisesi/neo/graphql/resolvers"
	"github.com/eveisesi/neo/graphql/service"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/killmail"
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
	r.Use(NewStructuredLogger(s.logger))
	r.Use(s.RateLimiter)
	r.Use(s.Dataloaders)

	schema := service.NewExecutableSchema(service.Config{
		Resolvers: &resolvers.Resolver{
			Services: resolvers.Services{
				Killmail:    s.killmail,
				Alliance:    s.alliance,
				Corporation: s.corporation,
				Character:   s.character,
				Universe:    s.universe,
			},
			Dataloader: CtxLoaders,
		},
	})

	r.Handle("/query", handler.GraphQL(
		schema,
		handler.IntrospectionEnabled(true),
	))

	r.Handle("/query/playground", handler.Playground(
		"GraphQL Playground",
		"/query",
	))

	r.Get("/auth/state", s.handleGetState)
	r.Post("/auth/token", s.handlePostCode)

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
