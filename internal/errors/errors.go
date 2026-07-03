package errors

import (
	"fmt"
)

// Exit codes (ADR 0001). Callers route on these without parsing text:
//
//	0  ok
//	1  operational error (has code + hint; safe to hand to an agent)
//	70 internal / unexpected (bug, panic; do not retry blindly)
//	80 usage / parse error (assigned by render() for Kong parse failures)
//
// ExitCode() never returns 80: the usage/parse case is synthesized in the
// render chokepoint, not derived from a CLIError code.
const (
	ExitSuccess   = 0
	ExitUserError = 1
	ExitInternal  = 70
	ExitUsage     = 80
)

// Error codes form a closed, documented set in snake_case, kept in Sol/Upsun
// domain vocabulary (never tuned to a specific caller). This is the API a
// machine caller branches on.
const (
	CodeUnauthenticated        = "unauthenticated"
	CodeNoProjectSpecified     = "no_project_specified"
	CodeNoEnvironmentSpecified = "no_environment_specified"
	CodeNotFound               = "not_found"
	CodeInvalidArgument        = "invalid_argument"
	CodePermissionDenied       = "permission_denied"
	CodeAPIUnavailable         = "api_unavailable"
	CodeOperationFailed        = "operation_failed"
	CodeInternal               = "internal"
)

// CLIError is Sol's structured error, shaped for machine consumption.
// Envelope: {code, message, hint?, retryable, details?}. code, message, and
// retryable are always present; hint and details are omitted when empty.
type CLIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Hint      string         `json:"hint,omitempty"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

func (e *CLIError) Error() string {
	return e.Message
}

// ExitCode returns the process exit code for this error. internal maps to 70;
// every other (operational) code maps to 1. The usage/parse code 80 is never
// returned here — render() assigns it for Kong parse failures.
func (e *CLIError) ExitCode() int {
	if e.Code == CodeInternal {
		return ExitInternal
	}
	return ExitUserError
}

// WithDetail adds a key/value to the error's details map.
func (e *CLIError) WithDetail(key string, value any) *CLIError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithHint sets the top-level hint field (an actionable next step).
func (e *CLIError) WithHint(hint string) *CLIError {
	e.Hint = hint
	return e
}

// WithRetryable marks whether the identical call may later succeed.
func (e *CLIError) WithRetryable(retryable bool) *CLIError {
	e.Retryable = retryable
	return e
}

// Common error constructors. Signatures are stable so call sites need no churn
// when the underlying code vocabulary evolves.

func NewAuthError(message string) *CLIError {
	err := &CLIError{
		Code:    CodeUnauthenticated,
		Message: message,
	}
	return err.WithHint("Run 'sol auth:login' to authenticate")
}

func NewAuthExpiredError() *CLIError {
	err := &CLIError{
		Code:    CodeUnauthenticated,
		Message: "Authentication expired and refresh failed",
	}
	return err.WithHint("Run 'sol auth:login' to re-authenticate")
}

func NewAuthMissingError() *CLIError {
	err := &CLIError{
		Code:    CodeUnauthenticated,
		Message: "Not authenticated",
	}
	return err.WithHint("Run 'sol auth:login' to authenticate")
}

// NewAPIError maps an HTTP status to the closest domain code. 401 is
// unauthenticated, 403 permission_denied, 404 not_found, 429 and 5xx
// api_unavailable (retryable); anything else is treated as an invalid request.
func NewAPIError(message string, statusCode int) *CLIError {
	err := &CLIError{
		Message: message,
		Details: map[string]any{"status_code": statusCode},
	}
	switch {
	case statusCode == 401:
		err.Code = CodeUnauthenticated
	case statusCode == 403:
		err.Code = CodePermissionDenied
	case statusCode == 404:
		err.Code = CodeNotFound
	case statusCode == 429:
		err.Code = CodeAPIUnavailable
		err.Retryable = true
		err.Hint = "Rate limited; wait and retry with reduced request frequency"
	case statusCode >= 500:
		err.Code = CodeAPIUnavailable
		err.Retryable = true
	default:
		err.Code = CodeInvalidArgument
	}
	return err
}

func NewNotFoundError(resource, id string) *CLIError {
	return &CLIError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found: %s", resource, id),
	}
}

func NewValidationError(message string) *CLIError {
	return &CLIError{
		Code:    CodeInvalidArgument,
		Message: message,
	}
}

// NewNoProjectError reports that no project could be resolved from flag or
// environment. Operational (exit 1) — the caller can recover by supplying one.
func NewNoProjectError() *CLIError {
	err := &CLIError{
		Code:    CodeNoProjectSpecified,
		Message: "no project specified",
	}
	return err.WithHint("Use --project or run from within a project directory")
}

// NewNoEnvironmentError reports that no environment could be resolved.
// Operational (exit 1) — the caller can recover by supplying one.
func NewNoEnvironmentError() *CLIError {
	err := &CLIError{
		Code:    CodeNoEnvironmentSpecified,
		Message: "no environment specified",
	}
	return err.WithHint("Provide an environment ID or use --environment flag")
}

// NewAPIUnreachableError reports a transport-level failure (DNS, refused
// connection, timeout): the request never got an HTTP answer. Operational and
// retryable — the identical call may succeed once connectivity recovers.
func NewAPIUnreachableError(message string) *CLIError {
	err := &CLIError{
		Code:      CodeAPIUnavailable,
		Message:   message,
		Retryable: true,
	}
	return err.WithHint("The Upsun API could not be reached; check connectivity and retry")
}

// NewOperationFailedError reports that a remote operation reached a non-success
// terminal state (failed, cancelled, or timed out while polling). Operational
// (exit 1), not an internal bug: the API answered correctly; the work itself
// did not succeed. Callers set Retryable via WithRetryable — a timeout may
// succeed on a later poll, but a failed or cancelled activity will not.
func NewOperationFailedError(message string) *CLIError {
	return &CLIError{
		Code:    CodeOperationFailed,
		Message: message,
	}
}

func NewInternalError(message string) *CLIError {
	err := &CLIError{
		Code:    CodeInternal,
		Message: message,
	}
	return err.WithHint("This is a bug. Please report it.")
}
