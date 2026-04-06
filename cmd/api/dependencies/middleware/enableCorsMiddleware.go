package middleware

import (
	"net/http"
)

func (a *MiddlewareDependencies) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range a.Config.Cors.TrustedOrigins {
				if origin == a.Config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Uncomment the next line if you want to allow credentials (cookies, etc.)
					// w.Header().Set("Access-Control-Allow-Credentials", "true")

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, PATCH, DELETE")
						requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
						if requestedHeaders != "" {
							w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
						} else {
							w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						}
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
