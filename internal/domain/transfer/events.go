package domain_transfer

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	CorrelationID() string
}

type TransferRequested struct {
	At             time.Time
	TransferID     uuid.UUID
	CorrelationID_ string

	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	AmountCents    int64
	Currency       string
	IdempotencyKey string
}

func (e TransferRequested) EventName() string { return "transfer.requested" }

func (e TransferRequested) OccurredAt() time.Time { return e.At }

func (e TransferRequested) AggregateID() uuid.UUID { return e.TransferID }

func (e TransferRequested) CorrelationID() string { return e.CorrelationID_ }

type TransferCompleted struct {
	At             time.Time
	TransferID     uuid.UUID
	CorrelationID_ string
	JournalID      uuid.UUID
}

func (e TransferCompleted) EventName() string { return "transfer.completed" }

func (e TransferCompleted) OccurredAt() time.Time { return e.At }

func (e TransferCompleted) AggregateID() uuid.UUID { return e.TransferID }

func (e TransferCompleted) CorrelationID() string { return e.CorrelationID_ }

type TransferFailed struct {
	At             time.Time
	TransferID     uuid.UUID
	CorrelationID_ string
	Reason         string
}

func (e TransferFailed) EventName() string { return "transfer.failed" }

func (e TransferFailed) OccurredAt() time.Time { return e.At }

func (e TransferFailed) AggregateID() uuid.UUID { return e.TransferID }

func (e TransferFailed) CorrelationID() string { return e.CorrelationID_ }
