// Package mocks provides mock implementations for testing purposes.
package mocks

//go:generate mockgen -destination=mock_persistence.go -package=mocks github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/persistence TransferRepository,OutboxRepository,InboxRepository,UnitOfWork
//go:generate mockgen -destination=mock_messaging.go -package=mocks github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/messaging Publisher
//go:generate mockgen -destination=mock_platform.go -package=mocks github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/platform Clock,IDGenerator
