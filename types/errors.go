package types

import "errors"

// Standard Sentinel Errors for the HotPlex SDK
var (
	// ErrDangerBlocked indicates that a command or input was intercepted and blocked by the WAF (Detector).
	ErrDangerBlocked = errors.New("danger event blocked: input matched forbidden patterns")

	// ErrInvalidConfig indicates that the provided configuration is invalid or missing required fields.
	ErrInvalidConfig = errors.New("invalid execution configuration")

	// ErrSessionNotFound indicates that the requested session does not exist in the pool.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionDead indicates that the session is no longer alive and cannot be used.
	ErrSessionDead = errors.New("session is dead")

	// ErrTimeout indicates that an operation timed out before completion.
	ErrTimeout = errors.New("operation timed out")

	// ErrInputTooLarge indicates that the input exceeds the maximum allowed size.
	ErrInputTooLarge = errors.New("input exceeds maximum allowed size")

	// ErrProcessStart indicates that the CLI process failed to start.
	ErrProcessStart = errors.New("failed to start CLI process")

	// ErrPipeClosed indicates that the pipe (stdin/stdout/stderr) is closed.
	ErrPipeClosed = errors.New("pipe is closed")
)
