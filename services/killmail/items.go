package killmail

import (
	"context"

	"github.com/eveisesi/neo"
)

func (s *service) ItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailItem, error) {
	return s.items.ByKillmailIDs(ctx, ids)
}
