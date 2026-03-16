package helpers

import "log/slog"

type HelperDependencies struct {
	Logger *slog.Logger
	// add other shared helper dependencies here, e.g., cache, config
}