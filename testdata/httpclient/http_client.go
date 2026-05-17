package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetBody fetches the body of a URL and returns it as a string.
// Uses a 10s timeout. Returns an error for non-200 status codes.
func GetBody(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("getBody: connecting to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getBody: server returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("getBody: reading body: %w", err)
	}
	return string(data), nil
}
