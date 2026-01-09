package messaging

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, exchange string, routingKey string, payload []byte, headers map[string]string) error
}
