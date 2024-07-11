package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or
			// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type any, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	//note thaat whatever is defined here before the return is only init once one the middleware is first created
	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	//this is a closure which will close over the limiter var
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//extract the clients ip address from the request
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		//lock the mutex to prevent this code from being executed concurrently
		mu.Lock()
		//check if the ip already exists in the map, if not init a new rate limiter and add the ip address and limiter to it
		if _, found := clients[ip]; !found {
			clients[ip] = rate.NewLimiter(2, 4)
		}
		//call the allow on the rate limiter for the current ip address, if not allowed unlock the mutex and send a 429
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		//very important, we must unlock the mutext before caling the next handler in the chain,
		//we DONT use defer here because that would mean the mutex isn't unlocked until all the handlers of this middleware have also returned
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
