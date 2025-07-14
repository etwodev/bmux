package router

// --- Internal structs ---

type route struct {
	id           int32
	name         string
	status       bool
	experimental bool
	handler      HandlerFunc
	middleware   []func(HandlerFunc) HandlerFunc
}

type router struct {
	status     bool
	routes     []Route
	middleware []func(HandlerFunc) HandlerFunc
}

// --- Route implementation ---

func (r route) ID() int32 {
	return r.id
}

func (r route) Status() bool {
	return r.status
}

func (r route) Experimental() bool {
	return r.experimental
}

func (r route) Handler() HandlerFunc {
	return r.handler
}

func (r route) Middleware() []func(HandlerFunc) HandlerFunc {
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

func (r router) Middleware() []func(HandlerFunc) HandlerFunc {
	return r.middleware
}

// --- Wrappers for extensibility ---

type RouterWrapper func(r Router) Router
type RouteWrapper func(r Route) Route

// --- Constructors ---

func NewRouter(
	status bool,
	routes []Route,
	middleware []func(HandlerFunc) HandlerFunc,
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
	id int32,
	status, experimental bool,
	handler HandlerFunc,
	middleware []func(HandlerFunc) HandlerFunc,
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
