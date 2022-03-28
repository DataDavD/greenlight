package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare an HTTP server using the same settings as in our main() function.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start a background goroutine.
	go func() {
		// Create a quit channel which carries os.Signal values. Use buffered
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay
		// them to the quit channel. Any other signal will not be caught by signal.Notify()
		// and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is
		// received.
		s := <-quit

		// Log a message to say that the signal has been caught. Notice that we also call the
		// String() method on the signal to get the signal name and include it in the log
		// entry properties.
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		// Exit the application with a 0 (success) status code.
		os.Exit(0)
	}()

	// Log a "starting server" message.
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Start the server as normal, returning any error.
	return srv.ListenAndServe()
}
