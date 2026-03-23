package handlers 

import (
	"log/slog"
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
)

type HandlerDependencies struct {
	Logger *slog.Logger
	Config dependencies.ServerConfig
	Helper helpers.HelperDependencies
	Models data.Models
}