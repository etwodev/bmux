# bmux

`bmux` is a modular TCP multiplexer and routing framework for Go. It provides a declarative interface for handling custom binary protocols using a router/middleware architecture inspired by modern web frameworks.

## Features

- **Pluggable Middleware Support** – Global, router-level, and route-level middleware chaining
- **Router-Based Dispatching** – Organize handlers in routers with isolated configuration
- **Binary Message Routing** – Parse and dispatch TCP messages based on custom headers
- **Concurrent Connection Handling** – Efficient handling of multiple clients
- **Structured Configuration** – Load runtime options via `config.Config`
- **Zerolog Integration** – Consistent and contextual structured logging support



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

// Define your header struct
type MyHeader struct {
	Command int32
}

func (h *MyHeader) ID() int32 {
	return h.Command
}

// Define a simple handler
func HelloHandler() router.HandlerFunc {
	return func(ctx *router.Context) {
		// handle the message
	}
}

func main() {
	s := bmux.New(&MyHeader{})

	// Create router and register routes
	r := router.New("main")
	r.Route(1, HelloHandler()) // ID 1 mapped to HelloHandler
	s.LoadRouter([]router.Router{r})

	// Start server with config.Address() and config.Port()
	if err := s.Start(); err != nil {
		panic(err)
	}
}
```

## Middleware

You can apply middleware at three levels:

* **Global**: Applies to all routes
* **Router-Level**: Applies to all routes within a router
* **Route-Level**: Applies to individual routes

### Example: Logging Middleware

```go
logger := log.NewZeroLogger(zerolog.New(os.Stdout).With().Timestamp().Logger())

loggingMiddleware := middleware.NewLoggingMiddleware(logger)
s.LoadMiddleware([]middleware.Middleware{loggingMiddleware})
```

Inside your route:

```go
if l := ctx.Context.Value(middleware.LoggerCtxKey); l != nil {
    if logger, ok := l.(log.Logger); ok {
        logger.Info().Msg("Processing route")
    }
}
```

## Configuration

`bmux` reads runtime configuration via the `config.Config` struct. This supports:

* Port and address binding
* Logging level (`debug`, `info`, `warn`, etc.)
* Timeout durations
* Keep-alive toggle
* Graceful shutdown duration
* Max concurrent connections

Ensure your `config.New()` call is invoked during startup to initialize values.

## Project Structure

```text
bmux/
├── bmux.go      → Core server and lifecycle
├── config/      → Config loading (JSON, env, etc.)
├── log/         → Abstraction over zerolog
├── middleware/  → Middleware primitives
├── parsing/     → TCP envelope and header parsing
├── router/      → Router, route, and context definitions
```

## Example Config File

```json
{
  "port": "9000",
  "address": "0.0.0.0",
  "logLevel": "debug",
  "bufferSize": 1024,
  "maxConnections": 100,
  "readTimeout": 10,
  "writeTimeout": 10,
  "idleTimeout": 30,
  "shutdownTimeout": 15,
  "enableKeepAlive": true,
  "enablePacketLogging": false
}
```

## Contributing

Contributions are welcome! Please:

1. Fork the repo
2. Create a feature branch
3. Write tests if applicable
4. Submit a PR with a clear description

## License

MIT License © 2025 [etwodev](https://github.com/etwodev)
