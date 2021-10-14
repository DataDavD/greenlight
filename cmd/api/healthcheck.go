package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in a response.
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request",
			http.StatusInternalServerError)
		return
	}

	// Append a newline to the JSON. This is just a small nicity to make it easier to view
	// in terminal applications.
	js = append(js, '\n')

	// At this point we know that the encoding the data worked without any problems, so we can
	// safely set any necessary HTTP headers for a successful response.
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON as the HTTP response body.
	if _, err := w.Write([]byte(js)); err != nil {
		app.logger.Println("error:", err)
		return
	}
}
