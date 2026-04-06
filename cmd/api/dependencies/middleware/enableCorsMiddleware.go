package middleware

import (
	"fmt"
	"net/http"
)

func (a *MiddlewareDependencies) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DEBUG: Print to stdout to confirm middleware is running
		fmt.Println("[CORS Middleware] Executed for:", r.Method, r.URL.Path, "Origin:", r.Header.Get("Origin"))
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		// Debug: log the incoming Origin header
		if a.Logger != nil {
			a.Logger.Info("CORS Debug: Incoming request", "Origin", r.Header.Get("Origin"), "Method", r.Method, "URL", r.URL.Path)
		}

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
						// Debug: log outgoing headers for preflight
						if a.Logger != nil {
							a.Logger.Info("CORS Debug: Preflight response headers", "Headers", w.Header())
						}
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		// Debug: log outgoing headers for normal requests
		if a.Logger != nil {
			a.Logger.Info("CORS Debug: Normal response headers", "Headers", w.Header())
		}

		next.ServeHTTP(w, r)
	})
}
