package config

// Config defines TCP/server-level configuration options.
type Config struct {
	Port                string `json:"port"`                // TCP listening port
	Address             string `json:"address"`             // TCP bind address
	Experimental        bool   `json:"experimental"`        // Enable experimental routes
	LogLevel            string `json:"logLevel"`            // Logging level (info, debug, etc.)
	BufferSize          int    `json:"bufferSize"`          // Size in bytes for read buffer
	MaxConnections      int    `json:"maxConnections"`      // Maximum simultaneous connections
	ReadTimeout         int    `json:"readTimeout"`         // Read timeout in seconds
	WriteTimeout        int    `json:"writeTimeout"`        // Write timeout in seconds
	IdleTimeout         int    `json:"idleTimeout"`         // Idle connection timeout in seconds
	ShutdownTimeout     int    `json:"shutdownTimeout"`     // Graceful shutdown timeout in seconds
	EnableKeepAlive     bool   `json:"enableKeepAlive"`     // Whether to enable TCP keep-alive
	EnablePacketLogging bool   `json:"enablePacketLogging"` // whether packet logging middleware should be enabled
}

func Port() string              { return c.Port }
func Address() string           { return c.Address }
func Experimental() bool        { return c.Experimental }
func LogLevel() string          { return c.LogLevel }
func BufferSize() int           { return c.BufferSize }
func MaxConnections() int       { return c.MaxConnections }
func ReadTimeout() int          { return c.ReadTimeout }
func WriteTimeout() int         { return c.WriteTimeout }
func IdleTimeout() int          { return c.IdleTimeout }
func ShutdownTimeout() int      { return c.ShutdownTimeout }
func EnableKeepAlive() bool     { return c.EnableKeepAlive }
func EnablePacketLogging() bool { return c.EnablePacketLogging }
