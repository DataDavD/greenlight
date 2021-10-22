package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Define an envelope type.
type envelope map[string]interface{}

// readIDParam reads interpolated "id" from request URL and returns it and nil. If there is an error
// it returns and 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope,
	headers http.Header) error {
	// use the json.MarshalIndent() function so that whitespace is added to the encoded JSON. Use
	// no line prefix and tab indents for each element.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// At this point, we know that we won't encounter any more errors before writing the response,
	// so it's safe to add any headers that we want to include. We loop through the header map
	// and add each header to the http.ResponseWriter header map. Note that it's OK if the
	// provided header map is nil. Go doesn't through an error if you try to range over (
	// or generally, read from) a nil map
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, the write the status code and JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(js)); err != nil {
		app.logger.Println("error:", err)
		return err
	}

	return nil
}
