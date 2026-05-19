package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBody(t *testing.T) {
	cases := []struct {
		name         string
		url          string
		responseCode int
		responseBody string
		expectError  bool
	}{
		{"successful get", "/success", http.StatusOK, "Hello, World!", false},
		{"non-200 status", "/notfound", http.StatusNotFound, "", true},
		{"read error", "/readerror", http.StatusOK, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var handler http.HandlerFunc
			if tc.name == "successful get" {
				handler = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.responseCode)
					fmt.Fprint(w, tc.responseBody)
				}
			} else if tc.name == "non-200 status" {
				handler = func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Not Found", tc.responseCode)
				}
			} else {
				handler = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.responseCode)
					_, _ = io.WriteString(w, "") // Simulating an error reading body
				}
			}

			ts := httptest.NewServer(handler)
			defer ts.Close()

			got, err := GetBody(ts.URL)
			if (err != nil) != tc.expectError {
				t.Errorf("GetBody(%q) unexpected error: %v", ts.URL, err)
				return
			}
			if !tc.expectError && got != tc.responseBody {
				t.Errorf("GetBody(%q) = %q; want %q", ts.URL, got, tc.responseBody)
			}
		})
	}
}