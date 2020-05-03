package killmail

import (
	"context"

	"github.com/eveisesi/neo"
)

func (s *service) AttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error) {
	return s.attackers.ByKillmailIDs(ctx, ids)
}
