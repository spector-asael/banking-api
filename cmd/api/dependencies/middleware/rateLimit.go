// filename: cmd/api/middleware.go

package middleware

import (
    "fmt"
    "net"
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

// RateLimit middleware enforces per-IP rate limiting with automatic stale client cleanup
func (a *MiddlewareDependencies) RateLimit(next http.Handler) http.Handler {
    // Define a rate limiter struct
    type client struct {
        limiter  *rate.Limiter
        lastSeen time.Time // remove map entries that are stale
    }

    var mu sync.Mutex                       // use to synchronize the map
    var clients = make(map[string]*client)  // the actual map

    // A goroutine to remove stale entries from the map
    go func() {
        for {
            time.Sleep(time.Minute)
            mu.Lock() // begin cleanup
            // delete any entry not seen in three minutes
            for ip, c := range clients {
                if time.Since(c.lastSeen) > 3*time.Minute {
                    delete(clients, ip)
                }
            }
            mu.Unlock() // finish cleanup
        }
    }()

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // get the IP address
        if a.Config.Limiter.Enabled {
            ip, _, err := net.SplitHostPort(r.RemoteAddr)
            if err != nil {
                a.Helpers.ServerErrorResponse(w, r, err)
                return
            }

            mu.Lock() // exclusive access to the map
            // check if ip address already in map, if not add it
            c, found := clients[ip]
            if !found {
                c = &client{
                    limiter:  rate.NewLimiter(rate.Limit(a.Config.Limiter.RPS), a.Config.Limiter.Burst),
                    lastSeen: time.Now(),
                }
                clients[ip] = c
            }

            // Update the last seen for the client
            c.lastSeen = time.Now()

            // Check the rate limit status
            if !c.limiter.Allow() {
                // Calculate dynamic Retry-After
                retryAfter := 1.0 / a.Config.Limiter.RPS

                mu.Unlock() // release lock before writing response

                w.Header().Set("Retry-After", fmt.Sprintf("%.2f seconds", retryAfter))
                a.Helpers.RateLimitExceededResponse(w, r)
                return
            }

            mu.Unlock() // others are free to get exclusive access to the map
        }

        next.ServeHTTP(w, r)
    })
}