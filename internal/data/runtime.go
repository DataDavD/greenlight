package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// MarshalJSON method on the Runtime type so that it satisfies the
// json.Marshaler interface. This should return the JSON-encoded string for the movie
// runtime in the format "<runtime> mins".
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it double quotes.
	// It needs to be surrounded by double quotes in order to be a valid JSON string.
	quotedJSONValue := strconv.Quote(jsonValue)

	// Convert the quoted string value to a byte slice and return it
	return []byte(quotedJSONValue), nil
}
