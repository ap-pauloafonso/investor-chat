package bot

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStockMessageWithServer(t *testing.T) {
	// Create a test server that serves a predefined CSV response
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve a CSV response with a valid stock quote
		if r.URL.Query().Get("s") == "AAPL" {
			response := `Symbol,Date,Time,Open,High,Low,Close,Volume
AAPL.US,2023-10-13,22:00:14,181.42,181.93,178.14,178.85,51456082`
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(response))
		} else if r.URL.Query().Get("s") == "GOOG" {
			// Serve a CSV response with "N/D" indicating an invalid stock code
			response := `Symbol,Date,Time,Open,High,Low,Close,Volume
ASDADADS,N/D,N/D,N/D,N/D,N/D,N/D,N/D`
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(response))
		} else {
			// Serve an empty response
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte{})
		}
	}))
	defer testServer.Close()

	t.Run("Valid Stock Quote", func(t *testing.T) {
		stockCode := "AAPL"
		message, err := GetStockMessage(testServer.URL+"?s=%s", stockCode)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedMessage := "AAPL.US quote is $178.85 per share"
		if message != expectedMessage {
			t.Errorf("Expected message: %s, got: %s", expectedMessage, message)
		}
	})

	t.Run("Invalid Stock Code", func(t *testing.T) {
		stockCode := "GOOG"
		message, err := GetStockMessage(testServer.URL+"?s=%s", stockCode)

		if err == nil {
			t.Errorf("Expected an error, but got none")
		}

		if !errors.Is(err, errInvalidStockCode) {
			t.Errorf("Expected error to be errInvalidStockCode, got %v", err)
		}

		if message != "" {
			t.Errorf("Expected an empty message, but got: %s", message)
		}
	})

	t.Run("Empty Response", func(t *testing.T) {
		stockCode := "INVALID"
		message, err := GetStockMessage(testServer.URL+"?s=%s", stockCode)

		if err == nil {
			t.Errorf("Expected an error, but got none")
		}

		if !errors.Is(err, errInvalidCsvFormat) {
			t.Errorf("Expected error to be errInvalidCsvFormat, got %v", err)
		}

		if message != "" {
			t.Errorf("Expected an empty message, but got: %s", message)
		}
	})
}
