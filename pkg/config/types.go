package config

// Config defines network-level configuration options.
type Config struct {
	Port            int    `json:"port"`            // Listening port (defaults to 30000)
	Protocol        string `json:"protocol"`        // What protocol to use (defaults to tcp://)
	Address         string `json:"address"`         // Bind address (defaults to 0.0.0.0)
	Experimental    bool   `json:"experimental"`    // Enable experimental routes (defaults to false)
	LogLevel        string `json:"logLevel"`        // Logging level (defaults to info)
	MaxConnections  int    `json:"maxConnections"`  // Maximum simultaneous connections (defaults to 1024)
	HeadSize        int    `json:"headSize"`        // The size of the header in bytes (defaults to 3)
	ShutdownTimeout int    `json:"shutdownTimeout"` // Graceful shutdown timeout in seconds (defaults to 15)
	EnableMulticore bool   `json:"enableMulticore"` // Whether to use multiple cores for the server (defaults to true)
}

func Port() int             { return c.Port }
func Protocol() string      { return c.Protocol }
func Address() string       { return c.Address }
func Experimental() bool    { return c.Experimental }
func LogLevel() string      { return c.LogLevel }
func MaxConnections() int   { return c.MaxConnections }
func HeadSize() int         { return c.HeadSize }
func ShutdownTimeout() int  { return c.ShutdownTimeout }
func EnableMulticore() bool { return c.EnableMulticore }
