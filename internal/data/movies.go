package data

import "time"

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
