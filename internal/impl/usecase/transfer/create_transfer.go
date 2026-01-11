package impl_transfer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domain_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/domain/transfer"
	port_persistence "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/persistence"
	port_platform "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/platform"
	port_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/usecase/transfer"
	"github.com/google/uuid"
)

type CreateTransferUsecaseImpl struct {
	uow   port_persistence.UnitOfWork
	repo  port_persistence.TransferRepository
	outbx port_persistence.OutboxRepository
	clock port_platform.Clock
	ids   port_platform.IDGenerator
}

func NewCreateTransferUsecaseImpl(
	uow port_persistence.UnitOfWork,
	repo port_persistence.TransferRepository,
	outbx port_persistence.OutboxRepository,
	clock port_platform.Clock,
	ids port_platform.IDGenerator,
) *CreateTransferUsecaseImpl {
	return &CreateTransferUsecaseImpl{
		uow:   uow,
		repo:  repo,
		outbx: outbx,
		clock: clock,
		ids:   ids,
	}
}

func (u *CreateTransferUsecaseImpl) Execute(ctx context.Context, in port_transfer.CreateTransferInput) (port_transfer.CreateTransferOutput, error) {
	if err := validateInput(in); err != nil {
		return port_transfer.CreateTransferOutput{}, err
	}

	requestHash := HashCreateTransferInput(in)

	existing, err := u.repo.GetByIdempotencyKey(ctx, in.IdempotencyKey)
	if err != nil && !errors.Is(err, port_persistence.ErrNotFound) {
		return port_transfer.CreateTransferOutput{}, fmt.Errorf("failed to check idempotency: %w", err)
	}

	if existing != nil {
		if existing.RequestHash != requestHash {
			return port_transfer.CreateTransferOutput{}, ErrIdempotencyConflict
		}

		return port_transfer.CreateTransferOutput{
			TransferID:    existing.Transfer.ID().String(),
			Status:        string(existing.Transfer.Status()),
			CreatedAt:     existing.Transfer.CreatedAt(),
			CorrelationID: existing.Transfer.CorrelationID(),
		}, nil
	}

	var transfer *domain_transfer.Transfer
	err = u.uow.WithinTx(ctx, func(txCtx context.Context) error {
		fromID, err := uuid.Parse(in.FromAccountID)
		if err != nil {
			return fmt.Errorf("%w: invalid from_account_id", ErrInvalidInput)
		}

		toID, err := uuid.Parse(in.ToAccountID)
		if err != nil {
			return fmt.Errorf("%w: invalid to_account_id", ErrInvalidInput)
		}

		now := u.clock.Now().UTC()
		transferID := u.ids.NewUUID()
		transfer, err = domain_transfer.New(domain_transfer.NewParams{
			TransferID:     transferID,
			FromAccountID:  fromID,
			ToAccountID:    toID,
			AmountCents:    in.AmountCents,
			Currency:       in.Currency,
			IdempotencyKey: in.IdempotencyKey,
			CorrelationID:  in.CorrelationID,
			Now:            now,
		})
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}

		if err := u.repo.Create(txCtx, transfer, requestHash); err != nil {
			return fmt.Errorf("failed to persist transfer: %w", err)
		}

		events := transfer.PullEvents()
		for _, event := range events {
			if err := u.enqueueEvent(txCtx, event, in.Traceparent); err != nil {
				return fmt.Errorf("failed to enqueue event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return port_transfer.CreateTransferOutput{}, err
	}

	return port_transfer.CreateTransferOutput{
		TransferID:    transfer.ID().String(),
		Status:        string(transfer.Status()),
		CreatedAt:     transfer.CreatedAt(),
		CorrelationID: transfer.CorrelationID(),
	}, nil
}

func validateInput(in port_transfer.CreateTransferInput) error {
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return fmt.Errorf("%w: missing idempotency_key", ErrInvalidInput)
	}
	if strings.TrimSpace(in.CorrelationID) == "" {
		return fmt.Errorf("%w: missing correlation_id", ErrInvalidInput)
	}
	if strings.TrimSpace(in.FromAccountID) == "" {
		return fmt.Errorf("%w: missing from_account_id", ErrInvalidInput)
	}
	if strings.TrimSpace(in.ToAccountID) == "" {
		return fmt.Errorf("%w: missing to_account_id", ErrInvalidInput)
	}
	if in.AmountCents <= 0 {
		return fmt.Errorf("%w: amount_cents must be > 0", ErrInvalidInput)
	}
	if len(strings.TrimSpace(in.Currency)) != 3 {
		return fmt.Errorf("%w: currency must be 3 letters", ErrInvalidInput)
	}
	return nil
}

func (u *CreateTransferUsecaseImpl) enqueueEvent(ctx context.Context, event domain_transfer.DomainEvent, traceparent string) error {
	messageID := u.ids.NewUUID()

	eventType := "TransferRequested"
	if req, ok := event.(domain_transfer.TransferRequested); ok {

		envelope := map[string]interface{}{
			"meta": map[string]interface{}{
				"schema_version": 1,
				"message_id":     messageID.String(),
				"event_type":     eventType,
				"producer":       "transfers-service",
				"correlation_id": event.CorrelationID(),
				"traceparent":    traceparent,
			},
			"data": map[string]interface{}{
				"transfer_id":     req.TransferID.String(),
				"from_account_id": req.FromAccountID.String(),
				"to_account_id":   req.ToAccountID.String(),
				"amount_cents":    req.AmountCents,
				"currency":        req.Currency,
			},
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		msg := port_persistence.OutboxMessage{
			MessageID:     messageID.String(),
			EventType:     eventType,
			AggregateType: "transfer",
			AggregateID:   event.AggregateID().String(),
			CorrelationID: event.CorrelationID(),
			Traceparent:   traceparent,
			Payload:       payload,
		}

		return u.outbx.Enqueue(ctx, msg)
	}

	return fmt.Errorf("unsupported event type: %T", event)
}
