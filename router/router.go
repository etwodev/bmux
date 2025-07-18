package router

import (
	"context"
	"net"
)

type Context struct {
	context.Context          // embeds standard context.Context
	Conn            net.Conn // TCP connection
	Header          any      // parsed header object
	Body            []byte   // message body
	MsgID           int      // message ID
}

type HandlerFunc func(*Context)

// Router defines a message-based router for the bmux protocol.
// It maps incoming message identifiers (int32) to handlers,
// supports middleware, and allows for enabling/disabling routers.
type Router interface {
	// Routes returns all registered routes in the router.
	Routes() []Route

	// Status indicates whether this router is currently active.
	Status() bool

	// Middleware returns router-level middleware applied to all routes.
	// Middleware wraps the handler with additional behavior.
	Middleware() []func(HandlerFunc) HandlerFunc
}

// Route defines a handler for a specific message ID in the bmux protocol.
type Route interface {
	// ID returns the int32 message ID this route handles.
	ID() int32

	// Name returns the name of the route, useful for logging.
	Name() string

	// Handler returns the bmux.HandlerFunc for this message.
	Handler() HandlerFunc

	// Status indicates whether the route is enabled.
	Status() bool

	// Experimental indicates if the route is experimental.
	Experimental() bool

	// Middleware returns middleware applied only to this route.
	Middleware() []func(HandlerFunc) HandlerFunc
}
