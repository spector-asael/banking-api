package dependencies 

import (
	"log/slog"
	"github.com/spector-asael/banking-api/internal/data"
)

type ServerConfig struct {
	Port int
	Environment  string
	DB struct {
        DSN string
    }
	Limiter struct {
        RPS float64                      // requests per second
        Burst int                        // initial requests possible
        Enabled bool                     // enable or disable rate limiter
    }
	Cors struct {
		TrustedOrigins []string
	}
}

type ApplicationDependencies struct {
	Config ServerConfig
	Logger *slog.Logger
	Models data.Models
}