package errors

import (
	"encoding/json"
	"fmt"
)

// Exit codes
const (
	ExitSuccess   = 0
	ExitUserError = 1 // Bad input, auth failed
	ExitAPIError  = 2 // API unreachable, server error
	ExitInternal  = 3 // Bug in CLI
)

// CLIError represents a structured error for agent consumption
type CLIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *CLIError) Error() string {
	return e.Message
}

// ExitCode returns the appropriate exit code for this error
func (e *CLIError) ExitCode() int {
	switch e.Code {
	case "AUTH_FAILED", "AUTH_EXPIRED", "AUTH_MISSING", "VALIDATION_ERROR", "NOT_FOUND":
		return ExitUserError
	case "API_ERROR", "RATE_LIMITED":
		return ExitAPIError
	case "INTERNAL_ERROR":
		return ExitInternal
	default:
		return ExitUserError
	}
}

// JSON returns the error as JSON bytes
func (e *CLIError) JSON() ([]byte, error) {
	wrapper := map[string]any{"error": e}
	return json.MarshalIndent(wrapper, "", "  ")
}

// WithDetail adds a detail to the error
func (e *CLIError) WithDetail(key string, value any) *CLIError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithHint adds a hint to the error details
func (e *CLIError) WithHint(hint string) *CLIError {
	return e.WithDetail("hint", hint)
}

// Common error constructors

func NewAuthError(message string) *CLIError {
	return &CLIError{
		Code:    "AUTH_FAILED",
		Message: message,
	}
}

func NewAuthExpiredError() *CLIError {
	err := &CLIError{
		Code:    "AUTH_EXPIRED",
		Message: "Authentication expired and refresh failed",
	}
	return err.WithHint("Run 'sol auth:login' to re-authenticate")
}

func NewAuthMissingError() *CLIError {
	err := &CLIError{
		Code:    "AUTH_MISSING",
		Message: "Not authenticated",
	}
	return err.WithHint("Run 'sol auth:login' to authenticate")
}

func NewAPIError(message string, statusCode int) *CLIError {
	return &CLIError{
		Code:    "API_ERROR",
		Message: message,
		Details: map[string]any{"status_code": statusCode},
	}
}

func NewNotFoundError(resource, id string) *CLIError {
	return &CLIError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found: %s", resource, id),
	}
}

func NewValidationError(message string) *CLIError {
	return &CLIError{
		Code:    "VALIDATION_ERROR",
		Message: message,
	}
}

func NewRateLimitError(retryAfter int) *CLIError {
	return &CLIError{
		Code:    "RATE_LIMITED",
		Message: "API rate limit exceeded",
		Details: map[string]any{
			"retry_after": retryAfter,
			"hint":        fmt.Sprintf("Wait %d seconds or reduce request frequency", retryAfter),
		},
	}
}

func NewInternalError(message string) *CLIError {
	err := &CLIError{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}
	return err.WithHint("This is a bug. Please report it.")
}
