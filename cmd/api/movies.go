package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DataDavD/snippetbox/greenlight/internal/data"
)

// createMovieHandler handles "POST /v1/movies" endpoint. For now, we just return a plain-text
// placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, "create a new movie")
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encounterd a problem and could not process your request",
			http.StatusInternalServerError)
		return
	}
}

// showMovieHandler handles "Get /v1/movies/:id" endpoint. For now, it returns plain-text
// placeholder response using the interpolated "id" parameter from the current URL
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, any interpolated URL Parameters will be stored
	// in the request context. We can use the ParamsFromContexnt() function to retrieve a slice
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
