package server_test

import (
	"http/v1/dev/internal/server"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomePage(t *testing.T) {
	t.Run("Test render homepage", func(t *testing.T) {
		testSrv, err := server.NewServer()
		if err != nil {
			t.Fatalf("Init server test failed: %q", err)
		}

		mockSrv := httptest.NewServer(testSrv.Handler())
		defer mockSrv.Close()

		resp, err := http.Get(mockSrv.URL + "/")
		if err != nil {
			t.Fatalf("Mock server test failed: %q", err)
		}

		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Mock server status with GET test failed, expected %q, got %q", http.StatusOK, resp.StatusCode)
		}

		if !strings.Contains(contentType, "text/html") {
			t.Errorf("Mock server GET content-type test failed, expected %q, got %q", "text/html", contentType)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Error reading the response body: %q", err)
		}

		if !strings.Contains(string(respBody), "Welcome, Guest") {
			t.Errorf("Incorrect response body, expected %q, got %q", "Welcome, Guest", string(respBody))
		}
	})
}
