package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Define a custom testServer type which anonymously embeds a httptest.Server instance.
type testServer struct {
	*httptest.Server
}

// newTestApp returns an instance of application struct
// containing mocked dependencies to be used for testing.
func newTestApp() *application {
	app := new(application)
	cfg := config{env: "testing"}
	app.config = cfg

	return app
}

// Create a newTestServer helper which initializes and returns a new instance of our
// custom testServer type.
func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	// Disable redirect-following for the client. Essentially this function is called
	// after a 3xx response is received by the client, and returning the http.ErrUseLastResponse
	// error forces it to immediately return the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

// Implement a get method on our custom testServer type. This makes a GET request to
// a given URL path on the test server, and returns the response status code, headers,
// and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := rs.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}
