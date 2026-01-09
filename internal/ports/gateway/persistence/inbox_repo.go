package port_persistence

import (
	"context"
)

type InboxRepository interface {
	TryInsert(ctx context.Context, consumer string, messageID string, messageType string, correlationID string) (bool, error)
	MarkProcessed(ctx context.Context, consumer string, messageID string) error
	MarkFailed(ctx context.Context, consumer string, messageID string, errMsg string) error
}
