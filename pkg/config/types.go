package config

// Config defines TCP-level configuration options.
type Config struct {
	Port            int    `json:"port"`            // TCP listening port (defaults to 30000)
	Address         string `json:"address"`         // TCP bind address (defaults to 0.0.0.0)
	Experimental    bool   `json:"experimental"`    // Enable experimental routes (defaults to false)
	LogLevel        string `json:"logLevel"`        // Logging level (defaults to info)
	MaxConnections  int    `json:"maxConnections"`  // Maximum simultaneous connections (defaults to 1024)
	HeadSize        int    `json:"headSize"`        // The size of the header in bytes (defaults to 3)
	ReadTimeout     int    `json:"readTimeout"`     // Read timeout in seconds (defaults to 15)
	ShutdownTimeout int    `json:"shutdownTimeout"` // Graceful shutdown timeout in seconds (defaults to 15)
	EnableKeepAlive bool   `json:"enableKeepAlive"` // Whether to enable TCP keep-alive (defaults to true)
	EnableMulticore bool   `json:"enableMulticore"` // Whether to use multiple cores for the server (defaults to true)
}

func Port() int             { return c.Port }
func Address() string       { return c.Address }
func Experimental() bool    { return c.Experimental }
func LogLevel() string      { return c.LogLevel }
func MaxConnections() int   { return c.MaxConnections }
func HeadSize() int         { return c.HeadSize }
func ReadTimeout() int      { return c.ReadTimeout }
func ShutdownTimeout() int  { return c.ShutdownTimeout }
func EnableKeepAlive() bool { return c.EnableKeepAlive }
func EnableMulticore() bool { return c.EnableMulticore }
