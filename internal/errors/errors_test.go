package errors

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		wantCode int
	}{
		// Operational errors (exit code 1)
		{"unauthenticated", &CLIError{Code: CodeUnauthenticated}, ExitUserError},
		{"no_project_specified", &CLIError{Code: CodeNoProjectSpecified}, ExitUserError},
		{"no_environment_specified", &CLIError{Code: CodeNoEnvironmentSpecified}, ExitUserError},
		{"not_found", &CLIError{Code: CodeNotFound}, ExitUserError},
		{"invalid_argument", &CLIError{Code: CodeInvalidArgument}, ExitUserError},
		{"permission_denied", &CLIError{Code: CodePermissionDenied}, ExitUserError},
		{"api_unavailable", &CLIError{Code: CodeAPIUnavailable}, ExitUserError},
		{"operation_failed", &CLIError{Code: CodeOperationFailed}, ExitUserError},

		// Internal errors (exit code 70)
		{"internal", &CLIError{Code: CodeInternal}, ExitInternal},

		// Unknown codes default to operational
		{"unknown code", &CLIError{Code: "made_up"}, ExitUserError},
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
		Code:    CodeInvalidArgument,
		Message: "something went wrong",
		Hint:    "try this instead",
		Details: map[string]any{"key": "value"},
	}

	data, jsonErr := err.JSON()
	if jsonErr != nil {
		t.Fatalf("JSON() returned error: %v", jsonErr)
	}

	var result map[string]any
	if parseErr := json.Unmarshal(data, &result); parseErr != nil {
		t.Fatalf("JSON() produced invalid JSON: %v", parseErr)
	}

	errorObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("JSON() missing 'error' wrapper")
	}

	if errorObj["code"] != CodeInvalidArgument {
		t.Errorf("code = %v, want %s", errorObj["code"], CodeInvalidArgument)
	}
	if errorObj["message"] != "something went wrong" {
		t.Errorf("message = %v, want 'something went wrong'", errorObj["message"])
	}
	if errorObj["hint"] != "try this instead" {
		t.Errorf("hint = %v, want 'try this instead'", errorObj["hint"])
	}
	// retryable is always present, even when false.
	if _, exists := errorObj["retryable"]; !exists {
		t.Error("JSON() should always include 'retryable'")
	}

	details, ok := errorObj["details"].(map[string]any)
	if !ok {
		t.Fatal("JSON() missing 'details'")
	}
	if details["key"] != "value" {
		t.Errorf("details.key = %v, want 'value'", details["key"])
	}
}

func TestJSONOmitsEmptyHintAndDetails(t *testing.T) {
	err := &CLIError{
		Code:    CodeInvalidArgument,
		Message: "no extras",
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
	if _, exists := errorObj["hint"]; exists {
		t.Error("JSON() should omit empty hint")
	}
}

func TestError(t *testing.T) {
	err := &CLIError{
		Code:    CodeInvalidArgument,
		Message: "the message",
	}

	if err.Error() != "the message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "the message")
	}
}

func TestWithDetail(t *testing.T) {
	err := &CLIError{Code: CodeInvalidArgument}

	result := err.WithDetail("foo", "bar")
	if result != err {
		t.Error("WithDetail should return the same error for chaining")
	}

	err.WithDetail("foo", "bar").WithDetail("num", 42)

	if err.Details["foo"] != "bar" {
		t.Errorf("Details[foo] = %v, want 'bar'", err.Details["foo"])
	}
	if err.Details["num"] != 42 {
		t.Errorf("Details[num] = %v, want 42", err.Details["num"])
	}
}

func TestWithHintSetsTopLevelField(t *testing.T) {
	err := &CLIError{Code: CodeInvalidArgument}
	result := err.WithHint("try this instead")

	if result != err {
		t.Error("WithHint should return the same error for chaining")
	}
	if err.Hint != "try this instead" {
		t.Errorf("Hint = %q, want 'try this instead'", err.Hint)
	}
	if _, exists := err.Details["hint"]; exists {
		t.Error("WithHint must not write details[hint]")
	}
}

func TestWithRetryable(t *testing.T) {
	err := &CLIError{Code: CodeAPIUnavailable}
	if err.WithRetryable(true); !err.Retryable {
		t.Error("WithRetryable(true) should set Retryable")
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		wantCode string
	}{
		{"NewAuthError", NewAuthError("failed"), CodeUnauthenticated},
		{"NewAuthExpiredError", NewAuthExpiredError(), CodeUnauthenticated},
		{"NewAuthMissingError", NewAuthMissingError(), CodeUnauthenticated},
		{"NewNotFoundError", NewNotFoundError("project", "123"), CodeNotFound},
		{"NewValidationError", NewValidationError("bad input"), CodeInvalidArgument},
		{"NewNoProjectError", NewNoProjectError(), CodeNoProjectSpecified},
		{"NewNoEnvironmentError", NewNoEnvironmentError(), CodeNoEnvironmentSpecified},
		{"NewRateLimitError", NewRateLimitError(60), CodeAPIUnavailable},
		{"NewOperationFailedError", NewOperationFailedError("deploy failed"), CodeOperationFailed},
		{"NewInternalError", NewInternalError("bug"), CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.wantCode)
			}
		})
	}
}

func TestNewAPIErrorMapsStatusToCode(t *testing.T) {
	tests := []struct {
		status        int
		wantCode      string
		wantRetryable bool
	}{
		{401, CodeUnauthenticated, false},
		{403, CodePermissionDenied, false},
		{404, CodeNotFound, false},
		{400, CodeInvalidArgument, false},
		{500, CodeAPIUnavailable, true},
		{503, CodeAPIUnavailable, true},
	}

	for _, tt := range tests {
		t.Run("status_"+strconv.Itoa(tt.status), func(t *testing.T) {
			err := NewAPIError("api error", tt.status)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", err.Code, tt.wantCode)
			}
			if err.Retryable != tt.wantRetryable {
				t.Errorf("Retryable = %v, want %v", err.Retryable, tt.wantRetryable)
			}
			if err.Details["status_code"] != tt.status {
				t.Errorf("Details[status_code] = %v, want %d", err.Details["status_code"], tt.status)
			}
		})
	}
}

func TestNewRateLimitErrorIsRetryable(t *testing.T) {
	err := NewRateLimitError(120)

	if !err.Retryable {
		t.Error("rate limit error should be retryable")
	}
	if err.Details["retry_after"] != 120 {
		t.Errorf("Details[retry_after] = %v, want 120", err.Details["retry_after"])
	}
	if err.Hint == "" {
		t.Error("rate limit error should carry a hint")
	}
}
