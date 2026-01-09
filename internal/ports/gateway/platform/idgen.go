package port_platform

import "github.com/google/uuid"

type IDGenerator interface {
	NewUUID() uuid.UUID
}
