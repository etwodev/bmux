# bmux

`bmux` is a modular TCP multiplexer and routing framework for Go. It provides a declarative interface for handling custom binary protocols using a router and middleware architecture inspired by modern web frameworks.

## Features

* Global, router-level, and route-level middleware chaining
* Organize handlers into routers with isolated configuration
* Parse and dispatch TCP messages using generic packet structures
* Built on top of the `gnet` async networking engine for efficient concurrency
* Load runtime options via `config.Config`

## Installation

```bash
go get github.com/etwodev/bmux
```

## Basic Usage

```go
import (
	"github.com/etwodev/bmux"
	"github.com/etwodev/bmux/router"
)

// Define your context struct
type MyContext struct {
	Command int32
}

// Define extraction functions as needed for your protocol

func main() {
	// Provide context factory and extraction functions to bmux.New
	contextFactory := func() *MyContext { return &MyContext{} }
	extractLength := func(c gnet.Conn, buf []byte) (headLen, totalLen int) { /* ... */ }
	extractMsgID := func(c gnet.Conn, head []byte) int { /* ... */ }

	server := bmux.New(contextFactory, extractLength, extractMsgID, nil)

	// Create router and register routes
	r := router.New("main")
	r.Route(1, router.HandlerFunc(func(ctx *router.Context) {
		// handle message with ID 1
	}))
	server.LoadRouter([]router.Router{r})

	// Start the server (listens on config.Address() and config.Port())
	server.Start()
}
```

## Middleware

Middleware can be applied at three levels:

* **Global** — applies to all routes
* **Router-Level** — applies to all routes within a router
* **Route-Level** — applies to individual routes

### Example: Logging Middleware

```go
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
loggingMiddleware := middleware.NewLoggingMiddleware(logger)
server.LoadMiddleware([]middleware.Middleware{loggingMiddleware})
```

Inside a handler:

```go
if l := ctx.Context.Value(middleware.LoggerCtxKey); l != nil {
    if logger, ok := l.(zerolog.Logger); ok {
        logger.Info().Msg("Processing route")
    }
}
```

## Configuration

`bmux` uses the `config.Config` struct to load runtime settings such as:

* Server address and port
* Logging level (e.g., `debug`, `info`, `warn`)
* Timeout durations (read, write, idle, shutdown)
* Maximum concurrent connections
* Enable or disable multi-core mode for `gnet`

Call `config.New()` early in your application to initialize the configuration.

## Project Structure

```
bmux/
├── bmux.go          → Core server and lifecycle management
├── pkg/config/          → Configuration loading and management
├── pkg/middleware/      → Middleware primitives and implementations
├── pkg/router/          → Router, route, and context definitions
├── pkg/engine/          → Core networking engine integration (gnet wrapper)
```

## Example Config File

```json
{
	"port": 9000,
	"address": "0.0.0.0",
	"experimental": false,
	"logLevel": "info",
	"maxConnections": 100,
	"readTimeout": 2,
	"shutdownTimeout": 15,
	"enableMulticore": true
}
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests when applicable
4. Submit a pull request with a clear description

## License

MIT License © 2025 [etwodev](https://github.com/etwodev)
