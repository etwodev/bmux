package bmux

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etwodev/bmux/pkg/config"
	"github.com/etwodev/bmux/pkg/engine"
	"github.com/etwodev/bmux/pkg/middleware"
	"github.com/etwodev/bmux/pkg/router"
	"github.com/panjf2000/gnet/v2"
	"github.com/rs/zerolog"
)

var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: "2006-01-02T15:04:05",
}).With().Timestamp().Str("Group", "bmux").Logger()

type Server[T any] struct {
	engineWrapper *engine.EngineWrapper[T]
	routers       []router.Router
	middleware    []middleware.Middleware
}

type Option[T any] func(*Server[T])

func New[T any](contextFactory func() *T, extractLength engine.ExtractLengthFunc[T], extractMsgID engine.ExtractMsgIDFunc[T], override *config.Config, opts ...Option[T]) *Server[T] {
	if contextFactory == nil {
		log.Fatal().Str("Function", "New").Msg("contextFactory cannot be nil")
	}

	if extractLength == nil {
		log.Fatal().Str("Function", "New").Msg("extractLength cannot be nil")
	}

	if extractMsgID == nil {
		log.Fatal().Str("Function", "New").Msg("extractMsgID cannot be nil")
	}

	if err := config.New(override); err != nil {
		log.Fatal().Str("Function", "New").Err(err).Msg("Failed to load config")
	}

	level, err := zerolog.ParseLevel(config.LogLevel())
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	engineWrapper := &engine.EngineWrapper[T]{
		ContextFactory: contextFactory,
		ExtractLength:  extractLength,
		ExtractMsgID:   extractMsgID,
		HeadSize:       config.HeadSize(),
	}

	s := &Server[T]{
		engineWrapper: engineWrapper,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server[T]) LoadRouter(routes []router.Router) {
	s.routers = append(s.routers, routes...)
}

func (s *Server[T]) LoadMiddleware(middleware []middleware.Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// registerRoutes registers routes and applies middleware chain for all routers and routes.
// This is called once at Start(), ensuring all routers and middleware are loaded.
func (s *Server[T]) registerRoutes() {
	for _, rt := range s.routers {
		if !rt.Status() {
			continue
		}

		for _, route := range rt.Routes() {
			if !route.Status() {
				continue
			}

			handler := route.Handler()

			// Route-level middleware (innermost) - wrapped first, so runs last
			for i := len(route.Middleware()) - 1; i >= 0; i-- {
				handler = route.Middleware()[i](handler)
			}

			// Router-level middleware
			for i := len(rt.Middleware()) - 1; i >= 0; i-- {
				handler = rt.Middleware()[i](handler)
			}

			// Global middleware
			for i := len(s.middleware) - 1; i >= 0; i-- {
				mw := s.middleware[i]
				if mw.Status() {
					handler = mw.Method()(handler)
				}
			}

			// Log route registration
			log.Debug().
				Str("Name", route.Name()).
				Int("RouteID", int(route.ID())).
				Bool("Experimental", route.Experimental()).
				Bool("Status", route.Status()).
				Msg("Registering route")

			// Register with engine wrapper
			s.engineWrapper.Handlers[route.ID()] = handler
		}
	}
}

func (s *Server[T]) Start() {
	s.registerRoutes()

	addr := fmt.Sprintf("%s:%d", config.Address(), config.Port())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		err := gnet.Run(s.engineWrapper, "tcp://"+addr, gnet.WithMulticore(config.EnableMulticore()))
		if err != nil {
			log.Fatal().Err(err).Msg("gnet server failed to start")
		}
		close(done)
	}()

	<-stop
	log.Warn().Msg("Interrupt received, initiating shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ShutdownTimeout())*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("error during graceful shutdown")
	}

	<-done
}

func (s *Server[T]) Shutdown(ctx context.Context) error {
	log.Warn().Str("Function", "Shutdown").Msg("Shutting down server")
	return s.engineWrapper.Engine.Stop(ctx)
}
