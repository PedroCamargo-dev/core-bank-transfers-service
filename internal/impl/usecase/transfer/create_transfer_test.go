package impl_transfer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domain_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/domain/transfer"
	impl_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/impl/usecase/transfer"
	gwmocks "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/mocks"
	port_persistence "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/persistence"
	port_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/usecase/transfer"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

const (
	testIdempotencyKey = "idem-123"
	testCorrelationID  = "corr-abc"
	testTraceparent    = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
)

type parsedEnvelope struct {
	Meta struct {
		SchemaVersion int    `json:"schema_version"`
		MessageID     string `json:"message_id"`
		EventType     string `json:"event_type"`
		Producer      string `json:"producer"`
		CorrelationID string `json:"correlation_id"`
		Traceparent   string `json:"traceparent"`
	} `json:"meta"`
	Data struct {
		TransferID    string `json:"transfer_id"`
		FromAccountID string `json:"from_account_id"`
		ToAccountID   string `json:"to_account_id"`
		AmountCents   int64  `json:"amount_cents"`
		Currency      string `json:"currency"`
	} `json:"data"`
}

func newService(ctrl *gomock.Controller) (*impl_transfer.CreateTransferUsecaseImpl,
	*gwmocks.MockUnitOfWork,
	*gwmocks.MockTransferRepository,
	*gwmocks.MockOutboxRepository,
	*gwmocks.MockClock,
	*gwmocks.MockIDGenerator,
) {
	uow := gwmocks.NewMockUnitOfWork(ctrl)
	repo := gwmocks.NewMockTransferRepository(ctrl)
	outbx := gwmocks.NewMockOutboxRepository(ctrl)
	clock := gwmocks.NewMockClock(ctrl)
	ids := gwmocks.NewMockIDGenerator(ctrl)

	svc := impl_transfer.NewCreateTransferUsecaseImpl(uow, repo, outbx, clock, ids)
	return svc, uow, repo, outbx, clock, ids
}

func TestCreateTransfer_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, uow, repo, outbx, clock, ids := newService(ctrl)

	repo.EXPECT().GetByIdempotencyKey(gomock.Any(), gomock.Any()).Times(0)
	uow.EXPECT().WithinTx(gomock.Any(), gomock.Any()).Times(0)
	outbx.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(0)
	clock.EXPECT().Now().Times(0)
	ids.EXPECT().NewUUID().Times(0)

	_, err := svc.Execute(context.Background(), port_transfer.CreateTransferInput{
		FromAccountID:  uuid.New().String(),
		ToAccountID:    uuid.New().String(),
		AmountCents:    1000,
		Currency:       "BRL",
		IdempotencyKey: "",
		CorrelationID:  testCorrelationID,
	})
	if !errors.Is(err, impl_transfer.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateTransfer_RepoError_Propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, uow, repo, outbx, clock, ids := newService(ctrl)

	repo.EXPECT().
		GetByIdempotencyKey(gomock.Any(), testIdempotencyKey).
		Return(nil, errors.New("db down"))

	uow.EXPECT().WithinTx(gomock.Any(), gomock.Any()).Times(0)
	outbx.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(0)
	clock.EXPECT().Now().Times(0)
	ids.EXPECT().NewUUID().Times(0)

	_, err := svc.Execute(context.Background(), port_transfer.CreateTransferInput{
		FromAccountID:  uuid.New().String(),
		ToAccountID:    uuid.New().String(),
		AmountCents:    1000,
		Currency:       "BRL",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateTransfer_Idempotent_ReturnsExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, uow, repo, outbx, clock, ids := newService(ctrl)

	now := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	fromID := uuid.New()
	toID := uuid.New()
	existingID := uuid.New()

	existing, _ := domain_transfer.New(domain_transfer.NewParams{
		TransferID:     existingID,
		FromAccountID:  fromID,
		ToAccountID:    toID,
		AmountCents:    1000,
		Currency:       "BRL",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
		Now:            now,
	})
	existing.PullEvents()

	in := port_transfer.CreateTransferInput{
		FromAccountID:  fromID.String(),
		ToAccountID:    toID.String(),
		AmountCents:    1000,
		Currency:       "brl",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
	}

	expectedHash := impl_transfer.HashCreateTransferInput(in)

	repo.EXPECT().
		GetByIdempotencyKey(gomock.Any(), testIdempotencyKey).
		Return(&port_persistence.StoredTransfer{
			Transfer:    existing,
			RequestHash: expectedHash,
		}, nil)

	uow.EXPECT().WithinTx(gomock.Any(), gomock.Any()).Times(0)
	outbx.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(0)
	clock.EXPECT().Now().Times(0)
	ids.EXPECT().NewUUID().Times(0)

	out, err := svc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.TransferID != existingID.String() {
		t.Fatalf("expected transfer_id %s, got %s", existingID, out.TransferID)
	}
}

func TestCreateTransfer_IdempotencyConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, uow, repo, outbx, clock, ids := newService(ctrl)

	now := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	fromID := uuid.New()
	toID := uuid.New()
	existingID := uuid.New()

	existing, _ := domain_transfer.New(domain_transfer.NewParams{
		TransferID:     existingID,
		FromAccountID:  fromID,
		ToAccountID:    toID,
		AmountCents:    1000,
		Currency:       "BRL",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
		Now:            now,
	})
	existing.PullEvents()

	in := port_transfer.CreateTransferInput{
		FromAccountID:  fromID.String(),
		ToAccountID:    toID.String(),
		AmountCents:    999,
		Currency:       "BRL",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
	}

	repo.EXPECT().
		GetByIdempotencyKey(gomock.Any(), testIdempotencyKey).
		Return(&port_persistence.StoredTransfer{
			Transfer:    existing,
			RequestHash: "original-hash",
		}, nil)

	uow.EXPECT().WithinTx(gomock.Any(), gomock.Any()).Times(0)
	outbx.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(0)
	clock.EXPECT().Now().Times(0)
	ids.EXPECT().NewUUID().Times(0)

	_, err := svc.Execute(context.Background(), in)
	if !errors.Is(err, impl_transfer.ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
}

func TestCreateTransfer_HappyPath_CreatesTransferAndOutbox(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, uow, repo, outbx, clock, ids := newService(ctrl)

	ctx := context.Background()

	now := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)

	fromID := uuid.New()
	toID := uuid.New()
	transferID := uuid.New()
	messageID := uuid.New()

	in := port_transfer.CreateTransferInput{
		FromAccountID:  fromID.String(),
		ToAccountID:    toID.String(),
		AmountCents:    1000,
		Currency:       "brl",
		IdempotencyKey: testIdempotencyKey,
		CorrelationID:  testCorrelationID,
		Traceparent:    testTraceparent,
	}

	repo.EXPECT().
		GetByIdempotencyKey(gomock.Any(), testIdempotencyKey).
		Return(nil, port_persistence.ErrNotFound)

	uow.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	clock.EXPECT().Now().Return(now)
	ids.EXPECT().NewUUID().Return(transferID)
	ids.EXPECT().NewUUID().Return(messageID)

	expectedHash := impl_transfer.HashCreateTransferInput(in)

	repo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, tr *domain_transfer.Transfer, requestHash string) error {
			if tr.ID() != transferID {
				t.Fatalf("expected transferID %s, got %s", transferID, tr.ID())
			}
			if tr.FromAccountID() != fromID || tr.ToAccountID() != toID {
				t.Fatalf("expected from/to ids to match input")
			}
			if tr.AmountCents() != 1000 {
				t.Fatalf("expected amount 1000, got %d", tr.AmountCents())
			}
			if tr.Currency() != "BRL" {
				t.Fatalf("expected currency BRL, got %s", tr.Currency())
			}
			if requestHash != expectedHash {
				t.Fatalf("expected requestHash %s, got %s", expectedHash, requestHash)
			}
			return nil
		})

	outbx.EXPECT().
		Enqueue(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, msg port_persistence.OutboxMessage) error {
			if msg.MessageID != messageID.String() {
				t.Fatalf("expected messageID %s, got %s", messageID, msg.MessageID)
			}
			if msg.EventType != "TransferRequested" {
				t.Fatalf("expected event type TransferRequested, got %s", msg.EventType)
			}
			if msg.CorrelationID != testCorrelationID {
				t.Fatalf("expected correlation_id %s, got %s", testCorrelationID, msg.CorrelationID)
			}

			var env parsedEnvelope
			if err := json.Unmarshal(msg.Payload, &env); err != nil {
				t.Fatalf("invalid payload json: %v", err)
			}
			if env.Meta.SchemaVersion != 1 {
				t.Fatalf("expected schema_version 1, got %d", env.Meta.SchemaVersion)
			}
			if env.Meta.EventType != "TransferRequested" {
				t.Fatalf("expected meta.event_type TransferRequested, got %s", env.Meta.EventType)
			}
			if env.Meta.CorrelationID != testCorrelationID {
				t.Fatalf("expected meta.correlation_id %s, got %s", testCorrelationID, env.Meta.CorrelationID)
			}
			if env.Meta.Traceparent != testTraceparent {
				t.Fatalf("expected traceparent to propagate")
			}
			if env.Data.TransferID != transferID.String() {
				t.Fatalf("expected data.transfer_id %s, got %s", transferID, env.Data.TransferID)
			}
			return nil
		})

	out, err := svc.Execute(ctx, in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.TransferID != transferID.String() {
		t.Fatalf("expected transfer_id %s, got %s", transferID, out.TransferID)
	}
	if out.Status != string(domain_transfer.StatusPending) {
		t.Fatalf("expected status PENDING, got %s", out.Status)
	}
	if out.CorrelationID != testCorrelationID {
		t.Fatalf("expected correlation %s, got %s", testCorrelationID, out.CorrelationID)
	}
}
