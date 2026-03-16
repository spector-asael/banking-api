package middleware

import (
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"log/slog"
)

type MiddlewareDependencies struct {
    Config dependencies.ServerConfig
    Logger *slog.Logger
    Helpers *helpers.HelperDependencies
}