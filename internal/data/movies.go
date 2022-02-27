package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DataDavD/snippetbox/greenlight/internal/validator"
	"github.com/lib/pq"
)

// Movie type whose fields describe the movie.
// Note that the Runtime type uses a custom Runtime type instead of int32. Furthermore, the omitempy
// directive on the Runtime type will still work on this: if the Runtime field has the underlying
// value 0, then it will be considered empty and omitted -- and the MarshalJSON() method won't
// be called.
type Movie struct {
	ID        int64     `json:"id"` // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`  // Use the - directive to never export in JSON output
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"` // Movie release year0
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"` // The version number starts at 1 and is incremented each
	// time the movie information is updated.
}

// MovieModel struct wraps a sql.DB connection pool and allows us to work with Movie struct type
// and the movies table in our database.
type MovieModel struct {
	DB *sql.DB
}

// Insert accepts a pointer to a movie struct, which should contain the data for the
// new record and inserts the record into the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from the movie
	// struct. Declaring this slice immediately next to our SQL query helps to make it nice and
	// clear *what values are being user where* in the query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get fetches a record from the movies table and returns the corresponding Movie struct.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts auto-incrementing
	// at 1 by default, so we know that no movies will have ID values less tan that.
	// To avoid making an unnecessary database call,
	// we take a shortcut and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Update the query to return pg_sleep(10) as the first value. This is to mimic a
	// long-running query.
	query := `
		SELECT PG_SLEEP(10), id, created_at, title, year, runtime, genres, version
        FROM movies
 		WHERE id = $1`

	// Declare a movie struct to hold the data returned by the query.
	var movie Movie

	// Importantly, update the Scan() parameters so taht the pg_sleep(10) return value
	// is scanned into a []byte slice.
	err := m.DB.QueryRow(query, id).Scan(
		&[]byte{},
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)

	// Handle any errors. If there was no matching movie found, Scan() will return a sql.ErrNoRows
	// error. We check for this and return our custom ErrRecordNotFound error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updates a specific movie in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // Add the expected movie version.
	}

	// Execute the SQL query. If no matching row could be found, we know the movie version
	// has changed (or the record has been deleted) and we return ErrEditConflict.
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete is a placeholder method for deleting a specific record in the movies table.
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1`

	// Execute the SQL query using the Exec() method,
	// passing in the id variable as the value for the placeholder parameter. The Exec(
	// ) method returns a sql.Result object.
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result
	// object to get the number of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected,
	// we know that the movies table didn't contain a record with the provided ID at the moment
	// we tried to delete it. In that case we return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// ValidateMovie runs validation checks on Movie type.
func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Check movie.Title
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	// Check input.Year
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// Check input.Runtime
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// Check input.Genres
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}
