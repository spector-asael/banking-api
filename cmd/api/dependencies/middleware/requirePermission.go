package middleware

import (
	"net/http"
)

// This middleware checks if the user has the right permissions
// We send the permission that is expected as an argument
func (a *MiddlewareDependencies) RequirePermission(permissionCode string,
	next http.HandlerFunc) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {
		user := a.Helpers.ContextGetUser(r)
		// get all the permissions associated with the user
		permissions, err := a.Models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			a.Helpers.ServerErrorResponse(w, r, err)
			return
		}
		if !permissions.Include(permissionCode) {
			a.Helpers.NotPermittedResponse(w, r)
			return
		}
		// they are good. Let's keep going
		next.ServeHTTP(w, r)
	}

	return a.RequireActivatedUser(fn)

}
