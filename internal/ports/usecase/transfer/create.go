package port_transfer

import (
	"context"
	"time"
)

type CreateTransferInput struct {
	FromAccountID  string
	ToAccountID    string
	AmountCents    int64
	Currency       string
	IdempotencyKey string
	CorrelationID  string
	Traceparent    string
}

type CreateTransferOutput struct {
	TransferID    string
	Status        string
	CreatedAt     time.Time
	CorrelationID string
}

type CreateTransferUseCase interface {
	Execute(ctx context.Context, input CreateTransferInput) (CreateTransferOutput, error)
}
