package domain_transfer

import "errors"

var (
	ErrInvalidTransferID     = errors.New("transfer: invalid transfer_id")
	ErrInvalidAccountID      = errors.New("transfer: invalid account_id")
	ErrSameAccount           = errors.New("transfer: from_account_id equals to_account_id")
	ErrInvalidAmount         = errors.New("transfer: amount_cents must be > 0")
	ErrInvalidCurrency       = errors.New("transfer: currency must be 3-letter ISO-like code")
	ErrMissingCorrelationID  = errors.New("transfer: correlation_id is required")
	ErrMissingIdempotencyKey = errors.New("transfer: idempotency_key is required")

	ErrInvalidStateTransition = errors.New("transfer: invalid state transition")
	ErrAlreadyFinalized       = errors.New("transfer: transfer already finalized")
	ErrMissingJournalID       = errors.New("transfer: journal_id is required to complete transfer")
	ErrMissingFailureReason   = errors.New("transfer: failure_reason is required to fail transfer")
)
