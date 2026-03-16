// filename: cmd/api/dependencies/errors/errors.go

package helpers

import (
  "fmt"
  "net/http"
  "errors"
)

var ErrRecordNotFound = errors.New("record not found")
// log an error message
func (a *HelperDependencies)logError(r *http.Request, err error)  {

   method := r.Method
   uri := r.URL.RequestURI()
   a.Logger.Error(err.Error(), "method", method, "uri", uri)
    
}

// send an error response if our server messes up
func (a *HelperDependencies)serverErrorResponse(w http.ResponseWriter,
                                                     r *http.Request, 
                                                     err error)  {

   // first thing is to log error message
   a.logError(r, err)
   // prepare a response to send to the client
   message := "the server encountered a problem and could not process your request"

   a.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}

// send an error response if our client messes up with a 404
func (a *HelperDependencies)notFoundResponse(w http.ResponseWriter,
                                                     r *http.Request)  {

   // we only log server errors, not client errors
   // prepare a response to send to the client
   message := "the requested resource could not be found"
   a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

// send an error response if our client messes up with a 405
func (a *HelperDependencies)methodNotAllowedResponse(
                                                     w http.ResponseWriter,
                                                     r *http.Request)  {

   // we only log server errors, not client errors
   // prepare a formatted response to send to the client
   message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
   a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

func (a *HelperDependencies)badRequestResponse(w http.ResponseWriter,
                                                     r *http.Request,
                                                     err error)  {

   a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *HelperDependencies)failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {

   a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}

// send an error response if rate limit exceeded (429 - Too Many Requests)
func (a *HelperDependencies)rateLimitExceededResponse(w http.ResponseWriter,
                                                        r *http.Request)  {

   message := "Error 429: rate limit exceeded"
   a.errorResponseJSON(w, r, http.StatusTooManyRequests, message)
}
