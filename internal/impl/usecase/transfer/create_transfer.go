package impl_transfer

import (
	"context"

	port_persistence "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/persistence"
	port_platform "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/gateway/platform"
	port_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/usecase/transfer"
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
	return port_transfer.CreateTransferOutput{}, ErrNotImplemented
}
