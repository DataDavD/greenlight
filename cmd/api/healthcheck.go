package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a fixed-format JSON response from a string. Notice how we're using a row string
	// literal (enclosed with backticks) so that we can include double-quote characters
	// in the JSON without needing to escape them.
	// We also use the %q verb to wrap the interpolated values in double-quotes
	js := `{"status": "available", "environment": %q, "version": %q}`
	js = fmt.Sprintf(js, app.config.env, version)

	// Set the "Content-Type: application/json" header on the response. If we forget to do this,
	// Go will default to sending a "Content-Type: text/plain; charset=utf-8" header instead.
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON as the HTTP response body.
	if _, err := w.Write([]byte(js)); err != nil {
		app.logger.Println("error:", err)
		return
	}
}
