package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetBody(t *testing.T) {
	cases := []struct {
		name           string
		url            string
		handler        http.HandlerFunc
		expectedBody   string
		expectedError  bool
	}{
		{
			name: "successful response",
			url:  "/success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "Hello, World!")
			},
			expectedBody:  "Hello, World!",
			expectedError: false,
		},
		{
			name: "404 response",
			url:  "/notfound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedBody:  "",
			expectedError: true,
		},
		{
			name: "error on connecting",
			url:  "/error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// simulate an internal server error
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedBody:  "",
			expectedError: true,
		},
		{
			name: "timeout",
			url:  "/timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(11 * time.Second) // longer than the client's timeout
			},
			expectedBody:  "",
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer ts.Close()

			body, err := GetBody(ts.URL + tc.url)

			if (err != nil) != tc.expectedError {
				t.Fatalf("GetBody() error = %v, wantErr %v", err, tc.expectedError)
			}
			if body != tc.expectedBody {
				t.Errorf("GetBody() = %v, want %v", body, tc.expectedBody)
			}
		})
	}
}