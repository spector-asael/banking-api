// Filename: cmd/api/dependencies/handlers/helpers.go
package helpers

import (
	"net/url"
	"strconv"
)

// Read string from query params
func (a *HelperDependencies) ReadString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

// Read int from query params
func (a *HelperDependencies) ReadInt(qs url.Values, key string, defaultValue int) int {
	s := qs.Get(key)

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return i
}