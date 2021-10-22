package data

import "time"

type Movie struct {
	ID        int64     // Unique integer ID for the movie
	CreatedAt time.Time // Timestamp for when the movie is added to our database
	Title     string
	Year      int32 // Movie release year
	Runtime   int32
	Genres    []string
	Version   int32 // The version number starts at 1 and will be incremented each time the movie
	// information is updated
}
