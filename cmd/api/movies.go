package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DataDavD/snippetbox/greenlight/internal/data"
)

// createMovieHandler handles "POST /v1/movies" endpoint. For now, we just return a plain-text
// placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP
	// request body (not that the field names and types in the struct are a subset of the Movie
	// struct). This struct will be our *target decode destination*.
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// Initialize a new json.Decoder instance which reads from teh request body, and then use the
	// Decode() method to decode the body contents into the input struct. Importantly,
	// notice that when we call Decode() we pass a *pointer* to the input struct as the target
	// decode destination. If there is an error during decoding, we use our generic errorResponse()
	// helper to send the client a 400 Bad Request response containing error message.
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Dump the contents of the input struct in an HTTP response.
	_, err = fmt.Fprintf(w, "%+v\n", input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// showMovieHandler handles "Get /v1/movies/:id" endpoint. For now, it returns plain-text
// placeholder response using the interpolated "id" parameter from the current URL
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, any interpolated URL Parameters will be stored
	// in the request context. We can use the ParamsFromContext() function to retrieve a slice
	// containing these paremter names and values.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from the URL and
	// some dummy data. Also notice that we deliberately haven't set a value for the Year field.
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Create an envelope{"movie": movie} instance and pass it to writeJSON(), instead of passing
	// the plain movie struct.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	return
}
