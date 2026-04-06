package dependencies

import (
	"log/slog"
	"sync"

	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/mailer"
)

type ServerConfig struct {
	Port        int
	Environment string
	DB          struct {
		DSN string
	}
	Limiter struct {
		RPS     float64 // requests per second
		Burst   int     // initial requests possible
		Enabled bool    // enable or disable rate limiter
	}
	Cors struct {
		TrustedOrigins []string
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

type ApplicationDependencies struct {
	Config ServerConfig
	Logger *slog.Logger
	Models data.Models
	Mailer mailer.Mailer
	WG     sync.WaitGroup // need this later for background jobs
}
