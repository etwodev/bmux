package middleware

import "github.com/etwodev/bmux/pkg/handler"

// Middleware defines the interface for bmux middleware that wraps bmux.HandlerFunc
// and provides metadata about the middleware such as name, status, and experimental flag.
//
// This interface enables middleware management, dynamic enabling/disabling, and identification.
type Middleware interface {
	// Method returns the middleware function that wraps a bmux.HandlerFunc.
	// The function signature is: func(bmux.HandlerFunc) bmux.HandlerFunc
	Method() func(handler.HandlerFunc) handler.HandlerFunc

	// Status returns true if the middleware is enabled, false otherwise.
	Status() bool

	// Experimental returns true if the middleware is experimental or unstable.
	Experimental() bool

	// Name returns the unique name of the middleware.
	Name() string
}
