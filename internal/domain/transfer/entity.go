package domain_transfer

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	id uuid.UUID

	fromAccountID uuid.UUID
	toAccountID   uuid.UUID
	amountCents   int64
	currency      string

	status         Status
	idempotencyKey string
	correlationID  string

	journalID     uuid.UUID
	failureReason string

	createdAt time.Time
	updatedAt time.Time

	pendingEvents []DomainEvent
}

type NewParams struct {
	TransferID     uuid.UUID
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	AmountCents    int64
	Currency       string
	IdempotencyKey string
	CorrelationID  string
	Now            time.Time
}

func New(p NewParams) (*Transfer, error) {
	if p.TransferID == uuid.Nil {
		return nil, ErrInvalidTransferID
	}

	if p.FromAccountID == uuid.Nil || p.ToAccountID == uuid.Nil {
		return nil, ErrInvalidAccountID
	}

	if p.FromAccountID == p.ToAccountID {
		return nil, ErrSameAccount
	}

	if p.AmountCents <= 0 {
		return nil, ErrInvalidAmount
	}

	cur := strings.ToUpper(strings.TrimSpace(p.Currency))
	if len(cur) != 3 {
		return nil, ErrInvalidCurrency
	}

	if strings.TrimSpace(p.IdempotencyKey) == "" {
		return nil, ErrMissingIdempotencyKey
	}

	if strings.TrimSpace(p.CorrelationID) == "" {
		return nil, ErrMissingCorrelationID
	}

	if p.Now.IsZero() {
		p.Now = time.Now().UTC()
	}

	t := &Transfer{
		id:             p.TransferID,
		fromAccountID:  p.FromAccountID,
		toAccountID:    p.ToAccountID,
		amountCents:    p.AmountCents,
		currency:       cur,
		status:         StatusPending,
		idempotencyKey: p.IdempotencyKey,
		correlationID:  p.CorrelationID,
		createdAt:      p.Now,
		updatedAt:      p.Now,
	}

	t.raise(TransferRequested{
		At:             p.Now,
		TransferID:     t.id,
		CorrelationID_: t.correlationID,
		FromAccountID:  t.fromAccountID,
		ToAccountID:    t.toAccountID,
		AmountCents:    t.amountCents,
		Currency:       t.currency,
		IdempotencyKey: t.idempotencyKey,
	})

	return t, nil
}

func (t *Transfer) Complete(journalID uuid.UUID, now time.Time) error {
	if t.status.IsFinal() {
		return ErrAlreadyFinalized
	}

	if t.status != StatusPending {
		return ErrInvalidStateTransition
	}

	if journalID == uuid.Nil {
		return ErrMissingJournalID
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.status = StatusCompleted
	t.journalID = journalID
	t.updatedAt = now

	t.raise(TransferCompleted{
		At:             now,
		TransferID:     t.id,
		CorrelationID_: t.correlationID,
		JournalID:      journalID,
	})

	return nil
}

func (t *Transfer) Fail(failureReason string, now time.Time) error {
	if t.status.IsFinal() {
		return ErrAlreadyFinalized
	}

	if t.status != StatusPending {
		return ErrInvalidStateTransition
	}

	failureReason = strings.TrimSpace(failureReason)
	if failureReason == "" {
		return ErrMissingFailureReason
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.status = StatusFailed
	t.failureReason = failureReason
	t.updatedAt = now

	t.raise(TransferFailed{
		At:             now,
		TransferID:     t.id,
		CorrelationID_: t.correlationID,
		Reason:         failureReason,
	})

	return nil
}

func (t *Transfer) PullEvents() []DomainEvent {
	if len(t.pendingEvents) == 0 {
		return nil
	}

	ev := make([]DomainEvent, len(t.pendingEvents))
	copy(ev, t.pendingEvents)

	t.pendingEvents = t.pendingEvents[:0]

	return ev
}

func (t *Transfer) raise(event DomainEvent) {
	t.pendingEvents = append(t.pendingEvents, event)
}

func (t *Transfer) ID() uuid.UUID { return t.id }

func (t *Transfer) FromAccountID() uuid.UUID { return t.fromAccountID }

func (t *Transfer) ToAccountID() uuid.UUID { return t.toAccountID }

func (t *Transfer) AmountCents() int64 { return t.amountCents }

func (t *Transfer) Currency() string { return t.currency }

func (t *Transfer) Status() Status { return t.status }

func (t *Transfer) IdempotencyKey() string { return t.idempotencyKey }

func (t *Transfer) CorrelationID() string { return t.correlationID }

func (t *Transfer) JournalID() uuid.UUID { return t.journalID }

func (t *Transfer) FailureReason() string { return t.failureReason }

func (t *Transfer) CreatedAt() time.Time { return t.createdAt }

func (t *Transfer) UpdatedAt() time.Time { return t.updatedAt }
