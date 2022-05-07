package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/DataDavD/snippetbox/greenlight/internal/data"
	"github.com/DataDavD/snippetbox/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

// recoverPanic is middleware that recovers from a panic by responding with a 500 Internal Server
// Error before closing the connection. It will also log the error using our custom Logger at
// the ERROR level.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic as
		// Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the response. This
				// acts a trigger to make Go's HTTP server automatically close the current
				// connection after a response has been sent.
				w.Header().Set("Connection:", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using our
				// custom Logger type at the ERROR level and send the client a
				// 500 Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for reach client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold pointers to a client struct.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once every
	// minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while the cleanup
			// is taking place.
			mu.Lock()

			// Loop through all clients. if they haven't been seen within the last three minutes,
			// then delete the corresponding entry from the clients map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limited is enabled.
		if app.config.limiter.enabled {
			// Extract the client's IP address from the request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't,
			// then initialize a new rate limiter and add teh IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				// Use the requests-per-second and burst values from the app.config struct.
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Call the limiter.Allow() method on the rate limiter for the current IP address.
			// If the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Very importantly, unlock the mutex before calling the next handler in the chain.
			// Notice that we DON'T use defer to unlock the mutex, as that would mean that the mutex
			// isn't unlocked until all handlers downstream of this middleware have also returned.
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any caches
		// that the response may vary based on the value of the Authorization header in the request.
		w.Header().Set("Vary", "Authorization")

		// Retrieve the value of the Authorization header from teh request. This will return the
		// empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header found, use the contextSetUser() helper to add
		// an AnonymousUser to the request context. Then we call the next handler in the chain
		// and return without executing any of the code below.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the header
		// isn't in the expected format we return a 401 Unauthorized response using the
		// invalidAuthenticationTokenResponse helper.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication toekn from the header parts
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format.
		v := validator.New()

		// If the token isn't valid, use the invalidAuthenticationtokenResponse
		// helper to send a response, rather than the failedValidatedResponse helper.
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrieve the details of the user associated with the authentication token.
		// call invalidAuthenticationTokenResponse if no matching record was found.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser healer to add the user information to the request context.
		r = app.contextSetUser(r, user)

		// Call next handler in chain
		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks that the user is not anonymous (i.e., they are authenticated).
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser helper to retrieve the user information from the request context.
		user := app.contextGetUser(r)

		// If the user is anonymous, then call authenticationRequiredResponse to inform the client
		// that they should be authenticated before trying again.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// requiredActivatedUser checks that the user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Rather than returning this http.HandlerFunc we assign it to the variable fn.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// Check that a user is activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requireAuthenticatedUser middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermissions(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context.
		user := app.contextGetUser(r)

		// Get the slice of permission for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Check if the slice includes the required permission. If it doesn't, then return a 403
		// Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Otherwise, they have the required permission so we call the next handler in the chain.
		next.ServeHTTP(w, r)
	})

	// Wrap this with the requireActivatedUser middleware before returning
	return app.requireActivatedUser(fn)
}
