package middleware

import (
  "net/http"
  "log"
)

func (a *MiddlewareDependencies)LoggingMiddleware(next http.Handler) http.Handler {
   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       log.Printf("Method: %s URL: %s", r.Method, r.URL.Path)
       next.ServeHTTP(w, r)
       log.Println("Request processed")
   })
}
