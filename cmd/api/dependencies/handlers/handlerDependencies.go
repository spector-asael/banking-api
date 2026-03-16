package handlers 

import (
	"log/slog"
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
)

type HandlerDependencies struct {
	Logger *slog.Logger
	Config dependencies.ServerConfig
	// Models data.Models
}