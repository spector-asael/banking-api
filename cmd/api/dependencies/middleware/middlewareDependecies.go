package middleware

import (
	"log/slog"

	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
)

type MiddlewareDependencies struct {
	Config  dependencies.ServerConfig
	Logger  *slog.Logger
	Helpers *helpers.HelperDependencies
	Models  *data.Models
}
