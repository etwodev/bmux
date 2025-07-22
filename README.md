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

// Define your context struct, this is what is stored with every request
type Context struct {
	IsEncrypted     bool
	ID              int32
}

// Define your get context wrapper
func GetContext() *Context { return &Context{IsEncrypted: false} }

// Define your "read length" extractor, for the following example, the packet is:
// | 0     | 1     | 2     | 3 ... 3+n-1        | 3+n ... 3+n+m-1     |
// |-------|-------|-------|--------------------|----------------------|
// | headLen | bodyLen (2 bytes LE) | header (n bytes) | body (m bytes) |
func GetReadLength() func(c gnet.Conn, buf []byte) (headLen int, totalLen int) {
	return func(c gnet.Conn, buf []byte) (headLen int, totalLen int) {
		if len(buf) < 3 {
			return 0, 0
		}

		headLen = int(buf[0])
		totalLen = headLen + int(binary.LittleEndian.Uint16(buf[1:3]))
		return headLen, totalLen
	}
}

// Define your "read head" extractor, this should extract the messageID or "identifier" from the packet
// You can also consume or set contextual information with Context.
func GetReadHead() func(c gnet.Conn, head []byte, body []byte) (msgID int) {
	return func(c gnet.Conn, head []byte, body []byte) (msgID int) {
		var h gen.MyProto

		if (ctx.IsEncrypted) {
			// Join head and body...
			// Decrypt payload
		}

		if err := proto.Unmarshal(head, &h); err != nil {
			return -1
		}

		ctx := c.Context().(*Context)
		ctx.ID = h.Msgid

		return int(h.Msgid)
	}
}



func main() {
	s := bmux.New(net.GetContext, net.GetReadLength(), net.GetReadHead(), nil)
	s.LoadRouter(Routers())
	s.LoadMiddleware(Middleware())
	s.Start()
}

// Group your routes
func Routers() []router.Router {
	return []router.Router{
		router.NewRouter(true, Routes(), nil),
	}
}

// Define your routes
func Routes() []router.Route {
	return []router.Route{
		router.NewRoute("Ping", 0x01, true, false, HandlePing(), nil),
	}
}


// Define your handler
func HandlePing() handler.HandlerFunc {
	return func(conn gnet.Conn, buf []byte) gnet.Action {
		var req gen.Ping
		if err := proto.Unmarshal(buf, &req); err != nil {
			fmt.Printf("Failed to unmarshal Ping: %v\n", err)
			return gnet.Close
		}

		body := gen.Ping{
			ServerTs: uint64(time.Now().UnixNano() / 1_000_000),
		}

		// ...Define your header
		// headByes := ...

		body, err := proto.Marshal(&body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}

		packet := make([]byte, 3+len(headBytes)+len(body))
		packet[0] = byte(len(headBytes))
		binary.LittleEndian.PutUint16(packet[1:3], uint16(len(body)))
		copy(packet[3:], headBytes)
		copy(packet[3+len(headBytes):], body)

		conn.Writev(packet)
		return gnet.None
	}
}

```

## Middleware

Middleware can be applied at three levels:

* **Global** — applies to all routes
* **Router-Level** — applies to all routes within a router
* **Route-Level** — applies to individual routes

### Example: Logging Middleware

```go
var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: "2006-01-02T15:04:05",
}).With().Timestamp().Str("Group", "bmux").Logger()

func Middleware(next handler.HandlerFunc) handler.HandlerFunc {
	return func(conn gnet.Conn, buf []byte) gnet.Action {
		ctx := conn.Context().(*net.Context)

		log.Info().
			Str("Group", "example-server").
			Int("MsgId", int(ctx.ID)).
			Str("Remote", conn.RemoteAddr().String()).
			Msg("Incoming message")

		return next(conn, buf)
	}
}
```

To register:

```go
func main() {
	// ...
	// s.LoadRouter(...)
	s.LoadMiddleware(Middleware())
	s.Start()
}

// ... func Routers()
// ...

func Middleware() []middleware.Middleware {
	return []middleware.Middleware{
		middleware.NewMiddleware(logging.Middleware, "connection_logger", true, true),
	}
}
```

## Configuration

`bmux` uses the `config.Config` struct to load runtime settings such as:

* Server address and port
* Logging level (e.g., `debug`, `info`, `warn`)
* Timeout duration (shutdown)
* Maximum concurrent connections
* Enable or disable multi-core mode for `gnet`

If you do not want to use the json config, you can set the config manually in bmux.New()

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
  "port": 30000,
  "address": "0.0.0.0",
  "experimental": false,
  "logLevel": "debug",
  "maxConnections": 1024,
  "headSize": 3,
  "shutdownTimeout": 10,
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
