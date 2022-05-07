package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes is our main application's router.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Convert the app.notFoundResponse helper to a http.Handler using the http.HandlerFunc()
	// adapter, and then set it as the custom error handler for 404 Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Convert app.methodNotAllowedResponse helper to a http.Handler and set it as the custom
	// error handler for 405 Method Not Allowed responses
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// Movies handlers. Note, that these movie endpoints use the `requireActivatedUser` middleware.
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermissions("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermissions("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermissions("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermissions("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermissions("movies:write", app.deleteMovieHandler))

	// Users handlers
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// Tokens handlers
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Wrap the router with the panic recovery middleware and rate limit middleware.
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
