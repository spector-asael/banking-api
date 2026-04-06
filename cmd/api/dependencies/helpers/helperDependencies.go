package helpers

import (
	"log/slog"
	"sync"
)

type HelperDependencies struct {
	Logger *slog.Logger
	WG     *sync.WaitGroup
}
