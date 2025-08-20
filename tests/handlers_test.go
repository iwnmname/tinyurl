package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tinyurl/internal/db"
	"tinyurl/internal/handlers"
	"tinyurl/internal/models"

	_ "modernc.org/sqlite"
)

var testServer *handlers.Server

func TestMain(m *testing.M) {
	database, err := initTestDB()
	if err != nil {
		panic("Failed to create test database: " + err.Error())
	}

	testServer = handlers.NewServer(database)

	exitCode := m.Run()

	database.Close()

	os.Exit(exitCode)
}

func TestShortenHandler(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		body           map[string]interface{}
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Valid URL",
			method:         http.MethodPost,
			body:           map[string]interface{}{"url": "https://example.com"},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Valid URL with TTL",
			method:         http.MethodPost,
			body:           map[string]interface{}{"url": "https://example.com", "ttl_days": 7},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Valid URL with Alias",
			method:         http.MethodPost,
			body:           map[string]interface{}{"url": "https://example.com", "alias": "test123"},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Empty URL",
			method:         http.MethodPost,
			body:           map[string]interface{}{"url": ""},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "Wrong Method",
			method:         http.MethodGet,
			body:           map[string]interface{}{"url": "https://example.com"},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req, err := http.NewRequest(tc.method, "/shorten", bytes.NewReader(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testServer.ShortenHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if tc.checkResponse {
				var response models.ShortenResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}
				if response.Code == "" {
					t.Errorf("Expected non-empty code in response")
				}
				if response.ShortURL == "" {
					t.Errorf("Expected non-empty shortURL in response")
				}
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	code := "redirect_test"
	url := "https://example.com"
	err := db.InsertLink(testServer.DB, code, url, 0)
	if err != nil {
		t.Fatalf("Failed to create test link: %v", err)
	}

	testCases := []struct {
		name           string
		code           string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid Code",
			code:           code,
			expectedStatus: http.StatusFound,
			expectedURL:    url,
		},
		{
			name:           "Nonexistent Code",
			code:           "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
		{
			name:           "Empty Code",
			code:           "",
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/r/"+tc.code, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testServer.RedirectHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if tc.expectedURL != "" {
				location := rr.Header().Get("Location")
				if location != tc.expectedURL {
					t.Errorf("handler returned wrong location: got %v want %v",
						location, tc.expectedURL)
				}
			}
		})
	}
}

func TestStatsHandler(t *testing.T) {
	code := "stats_test"
	url := "https://example.com"
	err := db.InsertLink(testServer.DB, code, url, 0)
	if err != nil {
		t.Fatalf("Failed to create test link: %v", err)
	}

	testCases := []struct {
		name           string
		code           string
		expectedStatus int
	}{
		{
			name:           "Valid Code",
			code:           code,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Nonexistent Code",
			code:           "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty Code",
			code:           "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/stats/"+tc.code, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(testServer.StatsHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response models.StatsResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}
				if response.URL != url {
					t.Errorf("Expected URL %s, got %s", url, response.URL)
				}
			}
		})
	}
}
