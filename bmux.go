package bmux

import (
	"context"
	"fmt"

	"net"
	"os"
	"reflect"
	"sync"
	"time"

	config "github.com/etwodev/bmux/config"
	"github.com/etwodev/bmux/log"
	"github.com/etwodev/bmux/middleware"
	"github.com/etwodev/bmux/parsing"
	"github.com/etwodev/bmux/router"
	"github.com/rs/zerolog"
)

// Server manages TCP connections and dispatches parsed messages
// to appropriate route handlers defined via the Router interface.
type Server struct {
	listener        net.Listener
	connections     map[net.Conn]struct{}
	mu              sync.Mutex
	wg              sync.WaitGroup
	quit            chan struct{}
	headerPrototype any                          // used to clone a new header per request
	routers         []router.Router              // routers loaded but not yet registered
	middlewares     []middleware.Middleware      // global middleware applied on all routes
	handlers        map[int32]router.HandlerFunc // flat route lookup for fast dispatch
	logger          log.Logger
}

// Option allows configuring the Server during creation.
type Option func(*Server)

// New returns a new Server instance with a required header prototype and optional settings.
func New(headerPrototype any, opts ...Option) *Server {
	if err := config.New(); err != nil {
		baseLogger := zerolog.New(os.Stdout).With().Timestamp().Str("Group", "bmux").Logger()
		baseLogger.Fatal().Str("Function", "New").Err(err).Msg("Failed to load config")
	}

	level, err := zerolog.ParseLevel(config.LogLevel())
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	format := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05"}
	baseLogger := zerolog.New(format).With().Timestamp().Str("Group", "bmux").Logger()

	logger := log.NewZeroLogger(baseLogger)

	if headerPrototype == nil {
		logger.Fatal().
			Str("Function", "New").
			Err(err).
			Msg("headerPrototype cannot be nil!")
	}

	s := &Server{
		connections:     make(map[net.Conn]struct{}),
		quit:            make(chan struct{}),
		headerPrototype: headerPrototype,
		handlers:        make(map[int32]router.HandlerFunc),
		logger:          logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// LoadRouter appends routers to the server's list, deferring route registration.
func (s *Server) LoadRouter(routers []router.Router) {
	s.routers = append(s.routers, routers...)
}

// LoadMiddleware appends global middleware to be applied to all routes.
func (s *Server) LoadMiddleware(mws []middleware.Middleware) {
	s.middlewares = append(s.middlewares, mws...)
}

// registerRoutes registers routes and applies middleware chain for all routers and routes.
// This is called once at Start(), ensuring all routers and middleware are loaded.
func (s *Server) registerRoutes() {
	for _, rt := range s.routers {
		if !rt.Status() {
			continue
		}

		for _, route := range rt.Routes() {
			if !route.Status() {
				continue
			}

			handler := route.Handler()

			// Apply route-level middleware (innermost) - RUNS LAST
			for i := len(route.Middleware()) - 1; i >= 0; i-- {
				handler = route.Middleware()[i](handler)
			}

			// Apply router-level middleware
			for i := len(rt.Middleware()) - 1; i >= 0; i-- {
				handler = rt.Middleware()[i](handler)
			}

			// Apply global middleware
			for i := len(s.middlewares) - 1; i >= 0; i-- {
				mw := s.middlewares[i]
				if mw.Status() {
					handler = mw.Method()(handler)
				}
			}

			// Apply root middleware (outermost) - RUNS FIRST
			if config.EnablePacketLogging() {
				handler = middleware.NewLoggingMiddleware(s.logger).Method()(handler)
			}

			// Log route registration
			s.logger.Debug().
				Str("Name", route.Name()).
				Int("RouteID", int(route.ID())).
				Bool("Experimental", route.Experimental()).
				Bool("Status", route.Status()).
				Msg("Registering route")

			s.handlers[route.ID()] = handler
		}
	}
}

// Start begins accepting connections on the configured TCP address and port.
func (s *Server) Start() error {
	s.registerRoutes()

	addr := net.JoinHostPort(config.Address(), config.Port())
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Debug().
		Str("Port", config.Port()).
		Str("Address", config.Address()).
		Bool("Experimental", config.Experimental()).
		Msg("Server starting")

	// If TCP listener supports keep-alive, enable it if configured
	if tcpListener, ok := s.listener.(*net.TCPListener); ok && config.EnableKeepAlive() {
		if err := tcpListener.SetDeadline(time.Time{}); err != nil {
			s.logger.Warn().
				Str("Function", "Start").
				Err(err).
				Msg("Failed to clear deadline for TCP keep-alive")
		}
	}

	go s.acceptLoop()

	<-s.quit

	s.logger.Warn().Str("Function", "Shutdown").Msg("Shutting server down...")
	_ = s.listener.Close()
	s.closeAllConnections()
	s.wg.Wait()

	return nil
}

// acceptLoop continuously accepts new connections and handles them.
func (s *Server) acceptLoop() {
	sem := make(chan struct{}, config.MaxConnections())

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				s.logger.Error().
					Str("Function", "acceptLoop").
					Err(err).
					Msg("Accept failed")

				continue
			}
		}

		sem <- struct{}{} // acquire slot

		s.mu.Lock()
		s.connections[conn] = struct{}{}
		s.mu.Unlock()

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			defer func() {
				s.mu.Lock()
				delete(s.connections, c)
				s.mu.Unlock()
				<-sem // release slot
				_ = c.Close()
			}()

			// Set timeouts on connection if configured
			if c, ok := c.(*net.TCPConn); ok {
				if rt := config.ReadTimeout(); rt > 0 {
					_ = c.SetReadDeadline(time.Now().Add(time.Duration(config.ReadTimeout()) * time.Second))
				}
				if wt := config.WriteTimeout(); wt > 0 {
					_ = c.SetWriteDeadline(time.Now().Add(time.Duration(config.WriteTimeout()) * time.Second))
				}
				if it := config.IdleTimeout(); it > 0 {
					_ = c.SetDeadline(time.Now().Add(time.Duration(config.IdleTimeout()) * time.Second))
				}
			}

			s.handleConnection(c)
		}(conn)
	}
}

// handleConnection reads messages from the connection and routes them.
func (s *Server) handleConnection(conn net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		_ = conn.Close()
	}()

	for {
		envelope, err := parsing.ParseEnvelope(conn)
		if err != nil {
			s.logger.Warn().
				Str("Function", "handleConnection").
				Err(err).
				Msg("Failed to parse envelope")
			break
		}

		header := s.newHeaderInstance()
		msgID, err := parsing.ParseHeader(envelope.RawHead, header)
		if err != nil {
			s.logger.Warn().
				Str("Function", "handleConnection").
				Err(err).
				Msg("Failed to parse header")
			continue
		}

		handler, ok := s.handlers[msgID]
		if !ok {
			s.logger.Warn().
				Int("MsgID", int(msgID)).
				Msg("No handler registered for message ID")
			continue
		}

		rctx := &router.Context{
			Context: ctx,
			Conn:    conn,
			Header:  header,
			Body:    envelope.RawBody,
			MsgID:   int(msgID),
		}

		handler(rctx)
	}
}

// closeAllConnections closes all active client connections.
func (s *Server) closeAllConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.connections {
		_ = conn.Close()
	}
}

// newHeaderInstance clones a new header instance from the provided prototype.
func (s *Server) newHeaderInstance() any {
	ptr := reflect.New(reflect.TypeOf(s.headerPrototype).Elem())
	return ptr.Interface()
}

// Shutdown initiates a graceful shutdown, closing the listener and waiting for all active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.quit)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
