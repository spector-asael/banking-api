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
}

type ApplicationDependencies struct {
	Config ServerConfig
	Logger *slog.Logger
	Models data.Models
}

func CreateDependencies() ApplicationDependencies {
	// 1. Create logger
	logger := slog.Default()

	// 2. Initialize your database / models
	// models := data.NewModels() // assuming you have a constructor for your Models

	// 3. Set up config
	config := ServerConfig{
		Port:        8080,
		Environment: "development",
	}
	config.DB.DSN = "user:password@tcp(localhost:3306)/banking"

	// 4. Return the dependencies container
	return ApplicationDependencies{
		Config: config,
		Logger: logger,
		// Models: models,
	}
}