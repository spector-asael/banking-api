// Filename: cmd/helpers/api/context.go
package helpers

import (
	"context"
	"net/http"

	"github.com/spector-asael/banking-api/internal/data"
)

// we need to create an alias for the 'user' key that we will add to
// the context object. The context object takes a key:value pair. However, the // context can be used by any other applications that our app interacts with // so we don't know if our 'user' key will be overwritten so we need to create // an alias to avoid that from happening
type contextKey string

const userContextKey = contextKey("user")

// Update the request context with the user information
// We return the request context with user-info added
func (a *HelperDependencies) ContextSetUser(r *http.Request,
	user *data.User) *http.Request {
	// WithValue() expects the original context along with the new
	// key:value pair you want to update it with
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Retrieve the user info when we expect it to be present (registered users)
// We can panic here because it means something went unexpectedly wrong
// .(*data.User) converts the value from a generic type (any) to a User type
func (a *HelperDependencies) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
