package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DataDavD/snippetbox/greenlight/internal/data"
	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	_ "github.com/lib/pq"
)

// Declare a string containing the application version number
const version = "1.0.0"

// Define a config struct.
type config struct {
	port int
	env  string
	// db struct field holds the configuration settings for our database connection pool.
	// For now this only holds the DSN, which we read in from a command-line flag.
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Define an application struct to hold dependencides for our HTTP handlers, helpers, and
// middleware.
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	models   data.Models
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Read the value of the port and env command-line flags into the config struct.
	// We default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production")

	// Read the DSN Value from the db-dsn command-line flag into the config struct.
	// We default to using our development DSN if no flag is provided.
	pw := os.Getenv("DB_PW")
	flag.StringVar(&cfg.db.dsn, "db-dsn",
		fmt.Sprintf("postgres://greenlight:%s@localhost/greenlight?sslmode=disable",
			pw), "PostgreSQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25,
		"PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25,
		"PostgreSQL max open idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m",
		"PostgreSQL max connection idle time")

	flag.Parse()

	// Initialize a new infoLog which writes messages to the STDOUT stream, prefixed
	// with the current date and time.
	infoLog := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Declare an instance of the application struct, containing the config struct and the infoLog.
	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	// Call the openDB() helper function (see below) to create teh connection pool,
	// passing in the config struct. If this returns an error,
	// we log it and exit the application immediately.
	db, err := openDB(cfg)
	if err != nil {
		app.errorLog.Fatal(err)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main()
	// function exits.
	func() {
		err := db.Close()
		if err != nil {
			app.errorLog.Fatal(err)
		}
	}()

	infoLog.Printf("database connection pool established")

	// Use the data.NewModels() function to add a Models struct to the application struct,
	// passing in the database connection pool as a parameter.
	app.models = data.NewModels(db)

	// Use the httprouter instance returned by app.routes as the server handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	app.infoLog.Printf("starting the %s server on %s", cfg.env, srv.Addr)
	// Because the "err" variable is now already declared in the code above,
	// we need to use the = operator here, instead of the := operator.
	if err = srv.ListenAndServe(); err != nil {
		app.errorLog.Fatal(err)
	}
}

// openDB returns a sql.DB connection pool to postgres database
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connection in the pool. Again,
	// passing a value less than or equal to 0 will mean there is no limit
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string to a
	// time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database,
	// passing in the context we created above as a parameter.
	// If connection couldn't be established successfully within the 5-second deadline,
	// then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
