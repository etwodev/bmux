package router

import "github.com/etwodev/bmux/pkg/handler"

// --- Internal structs ---

type route struct {
	id           int
	name         string
	status       bool
	experimental bool
	handler      handler.HandlerFunc
	middleware   []func(handler.HandlerFunc) handler.HandlerFunc
}

type router struct {
	status     bool
	routes     []Route
	middleware []func(handler.HandlerFunc) handler.HandlerFunc
}

// --- Route implementation ---

func (r route) ID() int {
	return r.id
}

func (r route) Status() bool {
	return r.status
}

func (r route) Experimental() bool {
	return r.experimental
}

func (r route) Handler() handler.HandlerFunc {
	return r.handler
}

func (r route) Middleware() []func(handler.HandlerFunc) handler.HandlerFunc {
	return r.middleware
}

func (r route) Name() string {
	return r.name
}

// --- Router implementation ---

func (r router) Routes() []Route {
	return r.routes
}

func (r router) Status() bool {
	return r.status
}

func (r router) Middleware() []func(handler.HandlerFunc) handler.HandlerFunc {
	return r.middleware
}

// --- Wrappers for extensibility ---

type RouterWrapper func(r Router) Router
type RouteWrapper func(r Route) Route

// --- Constructors ---

func NewRouter(
	status bool,
	routes []Route,
	middleware []func(handler.HandlerFunc) handler.HandlerFunc,
	opts ...RouterWrapper,
) Router {
	var r Router = router{
		status:     status,
		routes:     routes,
		middleware: middleware,
	}
	for _, o := range opts {
		r = o(r)
	}
	return r
}

func NewRoute(
	name string,
	id int,
	status, experimental bool,
	handler handler.HandlerFunc,
	middleware []func(handler.HandlerFunc) handler.HandlerFunc,
	opts ...RouteWrapper,
) Route {
	var r Route = route{
		name:         name,
		id:           id,
		status:       status,
		experimental: experimental,
		handler:      handler,
		middleware:   middleware,
	}
	for _, o := range opts {
		r = o(r)
	}
	return r
}
