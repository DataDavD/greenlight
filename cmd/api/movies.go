package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// createMovieHandler handles "POST /v1/movies" endpoint. For now, we just return a plain-text
// placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// showMovieHandler handles "Get /v1/movies/:id" endpoint. For now, it returns plain-text
// placeholder response using the interpolated "id" parameter from the current URL
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, any interpolated URL Parameters will be stored
	// in the request context. We can use the ParamsFromContexnt() function to retrieve a slice
	// containing these paremter names and values.
	params := httprouter.ParamsFromContext(r.Context())

	// We use the ByName() method to get the value of the "id" parameter from the slice.
	// In our project all movies will have a unique positive integer ID, but the value
	// returned ByName() is always a string. So, we try to convert it to a base 10 integer
	// (with a bit size of 64). If the paramter couldn't be converted, or is less than 1,
	// we know the ID is invalid so return the http.NotFound() function to return 404 Not Found
	// response.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// Otherwise, interpolate the movie ID in a placeholder response.
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}
