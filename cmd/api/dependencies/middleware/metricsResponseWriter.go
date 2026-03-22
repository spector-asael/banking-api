package middleware

import (
	"net/http"
)

// We will create a new type right before we use it in our metrics middleware
type metricsResponseWriter struct {
    wrapped    http.ResponseWriter   // the original http.ResponseWriter
    statusCode int         // this will contain the status code we need
    headerWritten bool    // has the response headers already been written?
}   

// Create an new instance of our custom http.ResponseWriter once
// we are provided with the original http.ResponseWriter. We will set
// the status code to 200 by default since that is what Golang does as well
// the headerWritten is false by default so no need to specify
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
    return &metricsResponseWriter {
        wrapped: w,
        statusCode: http.StatusOK,
    }
}


// Remember that the http.Header type is a map (key: value) of the headers
// Our custom http.ResponseWriter does not need to change the way the Header()
// method works, so all we do is call the original http.ResponseWriter's Header() 
// method when our custom http.ResponseWriter's Header() method is called
func (mw *metricsResponseWriter) Header() http.Header {
    return mw.wrapped.Header()
}
// Let's write the status code that is provided
// Again the original http.ResponseWriter's WriteHeader() methods knows
// how to do this
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
    mw.wrapped.WriteHeader(statusCode)
    // After the call to WriteHeader() returns, we record
    // the status code for use in our metrics

    // After the call to WriteHeader() returns, we record
    // the first status code for use in our metrics
    // NOTE: Because we only want the first status code sent, we will
    // ignore any other status code that gets written. For example,
    // mw.WriteHeader(404) followed by mw.WriteHeader(500). The client
    // will receive a 404, the 500 will never be sent
    if !mw.headerWritten {
        mw.statusCode = statusCode
        mw.headerWritten = true
    }
}
// The write() method simply calls the original http.ResponseWriter's
// Write() method which write the data to the connection
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
    mw.headerWritten = true
    return mw.wrapped.Write(b)
}
// We need a function to get the original http.ResponseWriter
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
    return mw.wrapped
}
