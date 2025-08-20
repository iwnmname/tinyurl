package tests

import (
	"crypto/tls"
	"net/http"
	"testing"

	"tinyurl/internal/utils"
)

func TestGenerateRandomCode(t *testing.T) {
	testCases := []struct {
		name   string
		length int
	}{
		{"Zero length", 0},
		{"Small length", 3},
		{"Normal length", 6},
		{"Large length", 12},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := utils.GenerateRandomCode(tc.length)
			if len(code) != tc.length {
				t.Errorf("GenerateRandomCode(%d) returned code with length %d, expected %d",
					tc.length, len(code), tc.length)
			}

			for _, char := range code {
				if !((char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9')) {
					t.Errorf("GenerateRandomCode returned invalid character: %c", char)
				}
			}
		})
	}
}

func TestGetHost(t *testing.T) {
	testCases := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name: "HTTP request",
			request: &http.Request{
				Host: "localhost:8080",
				TLS:  nil,
			},
			expected: "http://localhost:8080",
		},
		{
			name: "HTTPS request",
			request: &http.Request{
				Host: "example.com",
				TLS:  &tls.ConnectionState{},
			},
			expected: "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			host := utils.GetHost(tc.request)
			if host != tc.expected {
				t.Errorf("GetHost() returned %s, expected %s", host, tc.expected)
			}
		})
	}
}
