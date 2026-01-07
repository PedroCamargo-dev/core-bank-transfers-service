package domain_transfer_test

import (
	"errors"
	"testing"
	"time"

	domain_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/domain/transfer"
	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	validID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()
	now := time.Now().UTC()

	t.Run("creates transfer with valid parameters", func(t *testing.T) {
		transfer, err := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "USD",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            now,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.ID() != validID {
			t.Errorf("expected transfer id %v, got %v", validID, transfer.ID())
		}

		if transfer.FromAccountID() != fromAccountID {
			t.Errorf("expected from account id %v, got %v", fromAccountID, transfer.FromAccountID())
		}

		if transfer.ToAccountID() != toAccountID {
			t.Errorf("expected to account id %v, got %v", toAccountID, transfer.ToAccountID())
		}

		if transfer.AmountCents() != 1000 {
			t.Errorf("expected amount 1000, got %d", transfer.AmountCents())
		}

		if transfer.Currency() != "USD" {
			t.Errorf("expected currency USD, got %s", transfer.Currency())
		}

		if transfer.Status() != domain_transfer.StatusPending {
			t.Errorf("expected status pending, got %v", transfer.Status())
		}

		if transfer.IdempotencyKey() != "idempotency_key" {
			t.Errorf("expected idempotency key 'idempotency_key', got %s", transfer.IdempotencyKey())
		}

		if transfer.CorrelationID() != "correlation_id" {
			t.Errorf("expected correlation id 'correlation_id', got %s", transfer.CorrelationID())
		}

		if !transfer.CreatedAt().Equal(now) {
			t.Errorf("expected created at %v, got %v", now, transfer.CreatedAt())
		}

		if !transfer.UpdatedAt().Equal(now) {
			t.Errorf("expected updated at %v, got %v", now, transfer.UpdatedAt())
		}
	})

	t.Run("normalizes currency to uppercase", func(t *testing.T) {
		transfer, err := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "usd",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            now,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.Currency() != "USD" {
			t.Errorf("expected currency USD, got %s", transfer.Currency())
		}
	})

	t.Run("uses current time when Now is zero", func(t *testing.T) {
		transfer, err := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "USD",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            time.Time{},
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.CreatedAt().IsZero() {
			t.Error("expected created at to be set, got zero time")
		}

		if transfer.UpdatedAt().IsZero() {
			t.Error("expected updated at to be set, got zero time")
		}

		if !transfer.CreatedAt().Equal(transfer.UpdatedAt()) {
			t.Errorf("expected created at and updated at to be equal, got %v and %v", transfer.CreatedAt(), transfer.UpdatedAt())
		}
	})

	t.Run("raises TransferRequested event with correct data", func(t *testing.T) {
		transfer, _ := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "USD",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            now,
		})

		events := transfer.PullEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		event, ok := events[0].(domain_transfer.TransferRequested)
		if !ok {
			t.Fatalf("expected TransferRequested event, got %T", events[0])
		}

		if event.TransferID != validID {
			t.Errorf("expected transfer id %v, got %v", validID, event.TransferID)
		}

		if event.FromAccountID != fromAccountID {
			t.Errorf("expected from account id %v, got %v", fromAccountID, event.FromAccountID)
		}

		if event.ToAccountID != toAccountID {
			t.Errorf("expected to account id %v, got %v", toAccountID, event.ToAccountID)
		}

		if event.AmountCents != 1000 {
			t.Errorf("expected amount 1000, got %d", event.AmountCents)
		}

		if event.Currency != "USD" {
			t.Errorf("expected currency USD, got %s", event.Currency)
		}

		if event.IdempotencyKey != "idempotency_key" {
			t.Errorf("expected idempotency key 'idempotency_key', got %s", event.IdempotencyKey)
		}

		if event.CorrelationID() != "correlation_id" {
			t.Errorf("expected correlation id 'correlation_id', got %s", event.CorrelationID())
		}

		if !event.At.Equal(now) {
			t.Errorf("expected event occurred at %v, got %v", now, event.At)
		}

		eventsAgain := transfer.PullEvents()
		if len(eventsAgain) != 0 {
			t.Errorf("expected PullEvents to clear queue, got %d events on second call", len(eventsAgain))
		}
	})

	errorTests := []struct {
		name      string
		params    domain_transfer.NewParams
		wantError error
	}{
		{
			name: "returns error when transfer id is nil",
			params: domain_transfer.NewParams{
				TransferID:     uuid.Nil,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidTransferID,
		},
		{
			name: "returns error when from account id is nil",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  uuid.Nil,
				ToAccountID:    toAccountID,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidAccountID,
		},
		{
			name: "returns error when to account id is nil",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    uuid.Nil,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidAccountID,
		},
		{
			name: "returns error when from and to account ids are the same",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    fromAccountID,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrSameAccount,
		},
		{
			name: "returns error when amount is zero",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    0,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidAmount,
		},
		{
			name: "returns error when amount is negative",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    -1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidAmount,
		},
		{
			name: "returns error when currency is not 3 characters",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    1000,
				Currency:       "US",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrInvalidCurrency,
		},
		{
			name: "returns error when idempotency key is empty",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "",
				CorrelationID:  "correlation_id",
				Now:            now,
			},
			wantError: domain_transfer.ErrMissingIdempotencyKey,
		},
		{
			name: "returns error when correlation id is empty",
			params: domain_transfer.NewParams{
				TransferID:     validID,
				FromAccountID:  fromAccountID,
				ToAccountID:    toAccountID,
				AmountCents:    1000,
				Currency:       "USD",
				IdempotencyKey: "idempotency_key",
				CorrelationID:  "",
				Now:            now,
			},
			wantError: domain_transfer.ErrMissingCorrelationID,
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain_transfer.New(tt.params)

			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected error %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestTransfer_Complete(t *testing.T) {
	validID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()
	journalID := uuid.New()
	now := time.Now().UTC()
	laterTime := now.Add(time.Hour)

	createPendingTransfer := func() *domain_transfer.Transfer {
		transfer, _ := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "USD",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            now,
		})
		transfer.PullEvents()

		return transfer
	}

	t.Run("completes transfer with valid journal id", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Complete(journalID, laterTime)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.Status() != domain_transfer.StatusCompleted {
			t.Errorf("expected status completed, got %v", transfer.Status())
		}

		if transfer.JournalID() != journalID {
			t.Errorf("expected journal id %v, got %v", journalID, transfer.JournalID())
		}

		if !transfer.UpdatedAt().Equal(laterTime) {
			t.Errorf("expected updated at %v, got %v", laterTime, transfer.UpdatedAt())
		}
	})

	t.Run("uses current time when now is zero", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Complete(journalID, time.Time{})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.UpdatedAt().IsZero() {
			t.Error("expected updated at to be set, got zero time")
		}

		if !transfer.UpdatedAt().After(now) {
			t.Errorf("expected updated at to be after creation time %v, got %v", now, transfer.UpdatedAt())
		}
	})

	t.Run("raises TransferCompleted event with correct data", func(t *testing.T) {
		transfer := createPendingTransfer()

		transfer.Complete(journalID, laterTime)

		events := transfer.PullEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		event, ok := events[0].(domain_transfer.TransferCompleted)
		if !ok {
			t.Fatalf("expected TransferCompleted event, got %T", events[0])
		}

		if event.TransferID != validID {
			t.Errorf("expected transfer id %v, got %v", validID, event.TransferID)
		}

		if event.CorrelationID() != "correlation_id" {
			t.Errorf("expected correlation id 'correlation_id', got %s", event.CorrelationID())
		}

		if event.JournalID != journalID {
			t.Errorf("expected journal id %v, got %v", journalID, event.JournalID)
		}

		if !event.At.Equal(laterTime) {
			t.Errorf("expected event occurred at %v, got %v", laterTime, event.At)
		}
	})

	t.Run("returns error when journal id is nil", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Complete(uuid.Nil, laterTime)

		if !errors.Is(err, domain_transfer.ErrMissingJournalID) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrMissingJournalID, err)
		}

		if transfer.Status() != domain_transfer.StatusPending {
			t.Errorf("expected status to remain pending, got %v", transfer.Status())
		}
	})

	t.Run("returns error when transfer is already completed", func(t *testing.T) {
		transfer := createPendingTransfer()
		transfer.Complete(journalID, laterTime)

		err := transfer.Complete(uuid.New(), laterTime)

		if !errors.Is(err, domain_transfer.ErrAlreadyFinalized) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrAlreadyFinalized, err)
		}

		if transfer.JournalID() != journalID {
			t.Errorf("expected journal id to remain %v, got %v", journalID, transfer.JournalID())
		}
	})

	t.Run("returns error when transfer is already failed", func(t *testing.T) {
		transfer := createPendingTransfer()
		transfer.Fail("insufficient funds", laterTime)

		err := transfer.Complete(journalID, laterTime)

		if !errors.Is(err, domain_transfer.ErrAlreadyFinalized) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrAlreadyFinalized, err)
		}

		if transfer.Status() != domain_transfer.StatusFailed {
			t.Errorf("expected status to remain failed, got %v", transfer.Status())
		}
	})
}

func TestTransfer_Fail(t *testing.T) {
	validID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()
	now := time.Now().UTC()
	laterTime := now.Add(time.Hour)

	createPendingTransfer := func() *domain_transfer.Transfer {
		transfer, _ := domain_transfer.New(domain_transfer.NewParams{
			TransferID:     validID,
			FromAccountID:  fromAccountID,
			ToAccountID:    toAccountID,
			AmountCents:    1000,
			Currency:       "USD",
			IdempotencyKey: "idempotency_key",
			CorrelationID:  "correlation_id",
			Now:            now,
		})
		transfer.PullEvents()

		return transfer
	}

	t.Run("fails transfer with valid failure reason", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Fail("insufficient funds", laterTime)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.Status() != domain_transfer.StatusFailed {
			t.Errorf("expected status failed, got %v", transfer.Status())
		}

		if transfer.FailureReason() != "insufficient funds" {
			t.Errorf("expected failure reason 'insufficient funds', got %s", transfer.FailureReason())
		}

		if !transfer.UpdatedAt().Equal(laterTime) {
			t.Errorf("expected updated at %v, got %v", laterTime, transfer.UpdatedAt())
		}
	})

	t.Run("uses current time when now is zero", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Fail("insufficient funds", time.Time{})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.UpdatedAt().IsZero() {
			t.Error("expected updated at to be set, got zero time")
		}

		if !transfer.UpdatedAt().After(now) {
			t.Errorf("expected updated at to be after creation time %v, got %v", now, transfer.UpdatedAt())
		}
	})

	t.Run("trims whitespace from failure reason", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Fail("  insufficient funds  ", laterTime)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if transfer.FailureReason() != "insufficient funds" {
			t.Errorf("expected failure reason 'insufficient funds', got %s", transfer.FailureReason())
		}
	})

	t.Run("raises TransferFailed event with correct data", func(t *testing.T) {
		transfer := createPendingTransfer()

		transfer.Fail("insufficient funds", laterTime)

		events := transfer.PullEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		event, ok := events[0].(domain_transfer.TransferFailed)
		if !ok {
			t.Fatalf("expected TransferFailed event, got %T", events[0])
		}

		if event.TransferID != validID {
			t.Errorf("expected transfer id %v, got %v", validID, event.TransferID)
		}

		if event.CorrelationID() != "correlation_id" {
			t.Errorf("expected correlation id 'correlation_id', got %s", event.CorrelationID())
		}

		if event.Reason != "insufficient funds" {
			t.Errorf("expected reason 'insufficient funds', got %s", event.Reason)
		}

		if !event.At.Equal(laterTime) {
			t.Errorf("expected event occurred at %v, got %v", laterTime, event.At)
		}
	})

	t.Run("returns error when failure reason is empty", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Fail("", laterTime)

		if !errors.Is(err, domain_transfer.ErrMissingFailureReason) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrMissingFailureReason, err)
		}

		if transfer.Status() != domain_transfer.StatusPending {
			t.Errorf("expected status to remain pending, got %v", transfer.Status())
		}
	})

	t.Run("returns error when failure reason is only whitespace", func(t *testing.T) {
		transfer := createPendingTransfer()

		err := transfer.Fail("   ", laterTime)

		if !errors.Is(err, domain_transfer.ErrMissingFailureReason) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrMissingFailureReason, err)
		}

		if transfer.Status() != domain_transfer.StatusPending {
			t.Errorf("expected status to remain pending, got %v", transfer.Status())
		}
	})

	t.Run("returns error when transfer is already completed", func(t *testing.T) {
		transfer := createPendingTransfer()
		journalID := uuid.New()
		transfer.Complete(journalID, laterTime)

		err := transfer.Fail("insufficient funds", laterTime)

		if !errors.Is(err, domain_transfer.ErrAlreadyFinalized) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrAlreadyFinalized, err)
		}

		if transfer.Status() != domain_transfer.StatusCompleted {
			t.Errorf("expected status to remain completed, got %v", transfer.Status())
		}
	})

	t.Run("returns error when transfer is already failed", func(t *testing.T) {
		transfer := createPendingTransfer()
		transfer.Fail("first failure", laterTime)

		err := transfer.Fail("second failure", laterTime)

		if !errors.Is(err, domain_transfer.ErrAlreadyFinalized) {
			t.Errorf("expected error %v, got %v", domain_transfer.ErrAlreadyFinalized, err)
		}

		if transfer.FailureReason() != "first failure" {
			t.Errorf("expected failure reason to remain 'first failure', got %s", transfer.FailureReason())
		}
	})
}
