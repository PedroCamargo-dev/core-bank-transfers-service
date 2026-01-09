package port_persistence

import (
	"context"
	"errors"

	domain_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/domain/transfer"
)

var ErrNotFound = errors.New("persistence: not found")

type StoredTransfer struct {
	Transfer    *domain_transfer.Transfer
	RequestHash string
}

type TransferRepository interface {
	GetByIdempotencyKey(ctx context.Context, key string) (*StoredTransfer, error)
	Create(ctx context.Context, t *domain_transfer.Transfer, requestHash string) error
	GetByID(ctx context.Context, transferID string) (*StoredTransfer, error)
}
