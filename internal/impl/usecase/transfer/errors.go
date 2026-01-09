package impl_transfer

import "errors"

var (
	ErrIdempotencyConflict = errors.New("idempotency key conflict: different payload for same key")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrNotImplemented      = errors.New("not implemented yet")
)
