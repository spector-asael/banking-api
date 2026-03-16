package handlers 

import (
	"log/slog"
)

type HandlerDependencies struct {
	Logger *slog.Logger
}