package middleware

import (
	"context"

	"github.com/etwodev/bmux/log"
	"github.com/etwodev/bmux/router"
)

// NewLoggingMiddleware injects a logger into the request context.
func NewLoggingMiddleware(logger log.Logger) Middleware {
	return NewMiddleware(func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) {
			if ctx.Context == nil {
				panic("router.Context.Context is nil — middleware invoked before context was initialized")
			}

			ctx.Context = context.WithValue(ctx.Context, log.LoggerCtxKey, logger)
			next(ctx)
		}
	}, "inject_logger", true, true)
}
