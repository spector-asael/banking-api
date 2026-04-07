// Filename: cmd/api/helpers/middleware.go
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

func (a *MiddlewareDependencies) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This header tells the servers not to cache the response when
		// the Authorization header changes. This also means that the server is not
		// supposed to serve the same cached data to all users regardless of their
		// Authorization values. Each unique user gets their own cache entry
		w.Header().Add("Vary", "Authorization")

		// Get the Authorization header from the request. It should have the
		// Bearer token
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header then we have an Anonymous user
		if authorizationHeader == "" {
			r = a.Helpers.ContextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// Bearer token present so parse it. The Bearer token is in the form
		// Authorization: Bearer IEYZQUBEMPPAKPOAWTPV6YJ6RM
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			a.Helpers.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the actual token
		token := headerParts[1]
		// Validate
		v := validator.New()
		data.ValidateTokenPlaintext(v, token)
		if !v.IsEmpty() {
			a.Helpers.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the user info associated with this authentication token
		user, err := a.Models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				a.Helpers.InvalidAuthenticationTokenResponse(w, r)
			default:
				a.Helpers.ServerErrorResponse(w, r, err)
			}
			return
		}
		// Add the retrieved user info to the context
		r = a.Helpers.ContextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
