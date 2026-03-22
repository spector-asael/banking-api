package middleware 

import (
	"net/http"
	"expvar"
	"time"
	"strconv"
)

func (a *MiddlewareDependencies)MetricsMiddleware(next http.Handler) http.Handler {
	var (
		totalResponsesSentByStatus = expvar.NewMap("total_responses_sent_by_status")
		totalRequestsReceived = expvar.NewInt("total_requests_received")
		totalResponsesSent = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_seconds")
	) // Variables to track metrics 

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		startTime := time.Now() // Track the start time of the request processing
		totalRequestsReceived.Add(1) // Increment the total requests received counter

		// create a custom responseWriter
   		mw := newMetricsResponseWriter(w)

		// we send our custom responseWriter down the middleware chain
   		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1) // Increment the total responses sent counter after processing the request
		// extract the status code for use in our metrics since we have returned 
		// from the middleware chain. The map uses strings so we need to convert the
		// status codes from their integer values to strings
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
		duration := time.Since(startTime) // Calculate the duration of request processing
		totalProcessingTimeMicroseconds.Add(duration.Microseconds()) // Increment the total processing time counter with the duration of the request processing in microseconds
	})
}