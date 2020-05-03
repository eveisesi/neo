package killmail

import (
	"context"

	"github.com/eveisesi/neo"
)

func (s *service) VictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error) {
	return s.victim.ByKillmailIDs(ctx, ids)
}
