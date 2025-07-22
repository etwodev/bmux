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
	"github.com/etwodev/bmux/pkg/handler"
	"github.com/etwodev/bmux/pkg/middleware"
	"github.com/etwodev/bmux/pkg/router"
	"github.com/panjf2000/gnet/v2"
	"github.com/rs/zerolog"
)

var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: "2006-01-02T15:04:05",
}).With().Timestamp().Str("Group", "bmux").Logger()

// Server represents the bmux server instance.
// It manages routers, middleware, and the underlying event engine.
//
// T is a generic type parameter representing the connection context type.
//
// Usage:
//
//	ctxFactory := func() *MyContext { return &MyContext{} }
//	extractLen := func(c gnet.Conn, buf []byte) (headLen, totalLen int) { ... }
//	extractID := func(c gnet.Conn, head []byte) int { ... }
//
//	server := bmux.New(ctxFactory, extractLen, extractID, nil)
//	server.LoadRouter(myRouters)
//	server.LoadMiddleware(myMiddleware)
//	server.Start()
//
// The server handles connections using gnet for high-performance async I/O.
type Server[T any] struct {
	engineWrapper *engine.EngineWrapper[T]
	routers       []router.Router
	middleware    []middleware.Middleware
}

// Option defines a functional option to customize the Server.
type Option[T any] func(*Server[T])

// New creates a new bmux Server instance with the given context factory,
// length extractor, message ID extractor, optional config override, and options.
//
// It validates required arguments and loads configuration.
//
// Example:
//
//	ctxFactory := func() *MyContext { return &MyContext{} }
//	extractLen := func(c gnet.Conn, buf []byte) (headLen, totalLen int) { ... }
//	extractID := func(c gnet.Conn, head []byte) int { ... }
//
//	server := bmux.New(ctxFactory, extractLen, extractID, nil)
//
// The server is ready to have routers and middleware loaded before starting.
func New[T any](
	contextFactory func() *T,
	extractLength engine.ExtractLengthFunc[T],
	extractMsgID engine.ExtractMsgIDFunc[T],
	override *config.Config,
	opts ...Option[T],
) *Server[T] {
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
		MaxConnections: int64(config.MaxConnections()),
		Handlers:       make(map[int]handler.HandlerFunc),
	}

	s := &Server[T]{
		engineWrapper: engineWrapper,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// LoadRouter appends one or more routers to the server.
//
// Routers contain groups of routes and their associated middleware.
//
// Example:
//
//	myRouter := router.NewRouter(true, routes, middleware, opts...)
//	server.LoadRouter([]router.Router{myRouter})
func (s *Server[T]) LoadRouter(routes []router.Router) {
	s.routers = append(s.routers, routes...)
}

// LoadMiddleware appends global middleware to the server.
//
// Middleware applied here runs for all routes.
//
// Example:
//
//	server.LoadMiddleware([]middleware.Middleware{myMiddleware})
func (s *Server[T]) LoadMiddleware(middleware []middleware.Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// registerRoutes composes middleware chains and registers handlers
// from routers and routes into the engine's handler map.
//
// This method is invoked once automatically on server Start().
func (s *Server[T]) registerRoutes() {
	for _, rtr := range s.routers {
		if !rtr.Status() {
			continue
		}

		for _, rt := range rtr.Routes() {
			if !rt.Status() {
				continue
			}

			if rt.Experimental() && !config.Experimental() {
				continue
			}

			handler := rt.Handler()

			// Route-level middleware (innermost) - wrapped first, so runs last
			for i := len(rt.Middleware()) - 1; i >= 0; i-- {
				handler = rt.Middleware()[i](handler)
			}

			// Router-level middleware
			for i := len(rt.Middleware()) - 1; i >= 0; i-- {
				handler = rt.Middleware()[i](handler)
			}

			// Global middleware
			for i := len(s.middleware) - 1; i >= 0; i-- {
				mw := s.middleware[i]

				if !mw.Status() {
					continue
				}

				if mw.Experimental() && !config.Experimental() {
					continue
				}

				handler = mw.Method()(handler)
			}

			log.Debug().
				Str("Name", rt.Name()).
				Int("RouteID", int(rt.ID())).
				Bool("Experimental", rt.Experimental()).
				Bool("Status", rt.Status()).
				Msg("Registering route")

			s.engineWrapper.Handlers[rt.ID()] = handler
		}
	}
}

// Start launches the server, listening on the configured address and port,
// and gracefully handles shutdown on system interrupts.
//
// It blocks until the server exits.
//
// Example:
//
//	server.Start()
func (s *Server[T]) Start() {
	s.registerRoutes()

	addr := fmt.Sprintf("%s%s:%d", config.Protocol(), config.Address(), config.Port())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		err := gnet.Run(s.engineWrapper, addr, gnet.WithMulticore(config.EnableMulticore()))
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

// Shutdown gracefully stops the server using the provided context for timeout control.
//
// Returns any error encountered during shutdown.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	err := server.Shutdown(ctx)
func (s *Server[T]) Shutdown(ctx context.Context) error {
	log.Warn().Str("Function", "Shutdown").Msg("Shutting down server")
	return s.engineWrapper.Engine.Stop(ctx)
}
