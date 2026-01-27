package errors

import (
	"encoding/json"
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		wantCode int
	}{
		// User errors (exit code 1)
		{"AUTH_FAILED", &CLIError{Code: "AUTH_FAILED"}, ExitUserError},
		{"AUTH_EXPIRED", &CLIError{Code: "AUTH_EXPIRED"}, ExitUserError},
		{"AUTH_MISSING", &CLIError{Code: "AUTH_MISSING"}, ExitUserError},
		{"VALIDATION_ERROR", &CLIError{Code: "VALIDATION_ERROR"}, ExitUserError},
		{"NOT_FOUND", &CLIError{Code: "NOT_FOUND"}, ExitUserError},

		// API errors (exit code 2)
		{"API_ERROR", &CLIError{Code: "API_ERROR"}, ExitAPIError},
		{"RATE_LIMITED", &CLIError{Code: "RATE_LIMITED"}, ExitAPIError},

		// Internal errors (exit code 3)
		{"INTERNAL_ERROR", &CLIError{Code: "INTERNAL_ERROR"}, ExitInternal},

		// Unknown codes default to user error
		{"UNKNOWN_CODE", &CLIError{Code: "UNKNOWN_CODE"}, ExitUserError},
		{"empty code", &CLIError{Code: ""}, ExitUserError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ExitCode(); got != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	err := &CLIError{
		Code:    "TEST_ERROR",
		Message: "something went wrong",
		Details: map[string]any{"key": "value"},
	}

	data, jsonErr := err.JSON()
	if jsonErr != nil {
		t.Fatalf("JSON() returned error: %v", jsonErr)
	}

	// Parse the JSON to verify structure
	var result map[string]any
	if parseErr := json.Unmarshal(data, &result); parseErr != nil {
		t.Fatalf("JSON() produced invalid JSON: %v", parseErr)
	}

	// Verify wrapper structure
	errorObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("JSON() missing 'error' wrapper")
	}

	if errorObj["code"] != "TEST_ERROR" {
		t.Errorf("code = %v, want TEST_ERROR", errorObj["code"])
	}
	if errorObj["message"] != "something went wrong" {
		t.Errorf("message = %v, want 'something went wrong'", errorObj["message"])
	}

	details, ok := errorObj["details"].(map[string]any)
	if !ok {
		t.Fatal("JSON() missing 'details'")
	}
	if details["key"] != "value" {
		t.Errorf("details.key = %v, want 'value'", details["key"])
	}
}

func TestJSONOmitsEmptyDetails(t *testing.T) {
	err := &CLIError{
		Code:    "TEST_ERROR",
		Message: "no details",
	}

	data, jsonErr := err.JSON()
	if jsonErr != nil {
		t.Fatalf("JSON() returned error: %v", jsonErr)
	}

	var result map[string]any
	if parseErr := json.Unmarshal(data, &result); parseErr != nil {
		t.Fatalf("JSON() produced invalid JSON: %v", parseErr)
	}

	errorObj := result["error"].(map[string]any)
	if _, exists := errorObj["details"]; exists {
		t.Error("JSON() should omit empty details")
	}
}

func TestError(t *testing.T) {
	err := &CLIError{
		Code:    "TEST_ERROR",
		Message: "the message",
	}

	if err.Error() != "the message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "the message")
	}
}

func TestWithDetail(t *testing.T) {
	err := &CLIError{Code: "TEST"}
	err.WithDetail("foo", "bar").WithDetail("num", 42)

	if err.Details["foo"] != "bar" {
		t.Errorf("Details[foo] = %v, want 'bar'", err.Details["foo"])
	}
	if err.Details["num"] != 42 {
		t.Errorf("Details[num] = %v, want 42", err.Details["num"])
	}
}

func TestWithHint(t *testing.T) {
	err := &CLIError{Code: "TEST"}
	err.WithHint("try this instead")

	if err.Details["hint"] != "try this instead" {
		t.Errorf("Details[hint] = %v, want 'try this instead'", err.Details["hint"])
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		wantCode string
	}{
		{"NewAuthError", NewAuthError("failed"), "AUTH_FAILED"},
		{"NewAuthExpiredError", NewAuthExpiredError(), "AUTH_EXPIRED"},
		{"NewAuthMissingError", NewAuthMissingError(), "AUTH_MISSING"},
		{"NewAPIError", NewAPIError("api down", 500), "API_ERROR"},
		{"NewNotFoundError", NewNotFoundError("project", "123"), "NOT_FOUND"},
		{"NewValidationError", NewValidationError("bad input"), "VALIDATION_ERROR"},
		{"NewRateLimitError", NewRateLimitError(60), "RATE_LIMITED"},
		{"NewInternalError", NewInternalError("bug"), "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.wantCode)
			}
		})
	}
}

func TestNewAPIErrorIncludesStatusCode(t *testing.T) {
	err := NewAPIError("server error", 503)

	if err.Details["status_code"] != 503 {
		t.Errorf("Details[status_code] = %v, want 503", err.Details["status_code"])
	}
}

func TestNewRateLimitErrorIncludesRetryAfter(t *testing.T) {
	err := NewRateLimitError(120)

	if err.Details["retry_after"] != 120 {
		t.Errorf("Details[retry_after] = %v, want 120", err.Details["retry_after"])
	}
}
