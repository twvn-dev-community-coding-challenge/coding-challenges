// Package apperr defines the error categories shared by all modules.
// Callers wrap them with context (fmt.Errorf("%w: ...")) and the HTTP layer
// maps each category to a status code with errors.Is.
package apperr

import "errors"

var (
	ErrValidation        = errors.New("validation failed")
	ErrNotFound          = errors.New("not found")
	ErrInvalidTransition = errors.New("invalid transition")
	ErrConflict          = errors.New("conflict")

	ErrIdempotencyKeyMissing = errors.New("idempotency key missing")
	ErrIdempotencyKeyReused  = errors.New("idempotency key reused with a different request")
	ErrIdempotencyInProgress = errors.New("request with this idempotency key is in progress")
)
