package middleware

import (
	"net/http"
)

// This middleware checks if the user is authenticated (not anonymous)
func (a *MiddlewareDependencies) RequireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := a.Helpers.ContextGetUser(r)

		if user.IsAnonymous() {
			a.Helpers.AuthenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *MiddlewareDependencies) RequireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := a.Helpers.ContextGetUser(r)

		if !user.Activated {
			a.Helpers.InactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	// We pass the activation check middleware to the authentication
	// middleware to call (next) if the authentication check succeeds
	// In other words, only check if the user is activated if they are
	// actually authenticated.
	return a.RequireAuthenticatedUser(fn)
}
