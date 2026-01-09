package impl_transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	port_transfer "github.com/PedroCamargo-dev/core-bank-transfers-service/internal/ports/usecase/transfer"
)

func HashCreateTransferInput(in port_transfer.CreateTransferInput) string {
	from := strings.TrimSpace(in.FromAccountID)
	to := strings.TrimSpace(in.ToAccountID)
	cur := strings.ToUpper(strings.TrimSpace(in.Currency))

	payload := fmt.Sprintf("%s|%s|%d|%s", from, to, in.AmountCents, cur)

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
