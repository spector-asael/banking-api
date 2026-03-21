package middleware

import (
	"net/http"
)

func (a *MiddlewareDependencies) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin") 
		// Tell the browser not to cache the response if the Origin header is different
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range a.Config.Cors.TrustedOrigins {
				if origin == a.Config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// check if it is a Preflight CORS request
					if r.Method == http.MethodOptions && 
						r.Header.Get("Access-Control-Request-Method") != "" {
							w.Header().Set("Access-Control-Allow-Methods",
								"OPTIONS, PUT, PATCH, DELETE")
							w.Header().Set("Access-Control-Allow-Headers",
								"Authorization, Content-Type")
							// We need to send a 200 OK status code for Preflight CORS requests
							// Since this is a Preflight request, we can return early and not call the next handler
							w.WriteHeader(http.StatusOK)
							return
						}
					break
				}
			}
		}

		next.ServeHTTP(w,r)
	})
}