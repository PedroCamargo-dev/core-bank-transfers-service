package domain_transfer

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

func (s Status) IsFinal() bool {
	return s == StatusCompleted || s == StatusFailed
}
