package port_persistence

import "context"

type OutboxMessage struct {
	MessageID     string
	EventType     string
	AggregateType string
	AggregateID   string
	CorrelationID string
	Traceparent   string
	Payload       []byte
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, msg OutboxMessage) error
	DequeueBatch(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkPublished(ctx context.Context, messageID string) error
}
