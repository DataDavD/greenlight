package main

import (
	"fmt"
	"net/http"
)

// logError method is a generic helper for logging an error message in *application.
// Right now it's a basic infoLog.Println, but we'll upgrade this later to
// use structured logging and record additional info about the request.
func (app *application) logError(r *http.Request, err error) {
	app.infoLog.Println(err)
}

// errorResponse method is a generic helper for sending JSON-formatted error messages to the
// client with a given status code. Note that we're using an interface{} type for the message
// parameter, rather than just a string type, as this gives us more flexibility over the values
// that we can include in the response.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	// Write the response using the writeJSON() helper. If this happens to return an error
	// then log it, and fall back to sending the client an empty response with a 500 Internal
	// Server Error status code
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse method is used when our application encounters an unexpected problem
// at runtime. it logs the detailed error message, then uses the errorResponse() helper to send a
// 500 Internal Server Error status code and JSON response (containing the generic error message)
// to the client
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, 500, message)
}

// notFoundResponse method is used to send a 404 Not Found status code and JSON response to the
// client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse method is used to send a 405 Method Not Allowed status code and
// JSON response to the client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// badRequestResponse send JSON-formatted error message with 400 Bad Request status code.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse send JSON-formatted error message to client with given status code
// when Validation fails.
// Note that the errors parameter here has the type map[string]string,
// which is exact the same as the errors map contained in our Validator type.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
