package middleware

import (
   "fmt"
   "net/http"
)

func (a *MiddlewareDependencies) RecoverPanic(next http.Handler) http.Handler  {

   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
   // defer will be called when the stack unwinds
       defer func() {
           // recover() checks for panics
           err := recover();
           if err != nil {
               w.Header().Set("Connection", "close")
               a.Helpers.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
           }
       }()
       next.ServeHTTP(w,r)
   })  
}