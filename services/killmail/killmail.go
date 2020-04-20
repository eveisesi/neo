package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

var rkillmail = "killmail:%d:%s"

func (s *service) Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)
	var key = fmt.Sprintf(rkillmail, id, hash)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, killmail)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal killmail from redis")
		}
		return killmail, nil
	}

	killmail, err = s.KillmailRespository.Killmail(ctx, id, hash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for killmail")
	}

	byteSlc, err := json.Marshal(killmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal killmail for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Minute*60).Result()

	return killmail, errors.Wrap(err, "failed to cache killmail in redis")

}

// func (s *service) KillmailsByCharacterID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {
// 	return s.KillmailRespository.KillmailsByCharacterID(ctx, id)
// }

// func (s *service) KillmailsByCorporationID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {
// 	return s.KillmailRespository.KillmailsByCorporationID(ctx, id)
// }

// func (s *service) KillmailsByAllianceID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {
// 	return s.KillmailRespository.KillmailsByAllianceID(ctx, id)
// }

// func (s *service) KillmailsByFactionID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {
// 	return s.KillmailRespository.KillmailsByFactionID(ctx, id)
// }

// func (s *service) KillmailAttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error) {
// 	return s.KillmailRespository.KillmailAttackersByKillmailIDs(ctx, ids)
// }

// func (s *service) KillmailItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailItem, error) {
// 	return s.KillmailRespository.KillmailItemsByKillmailIDs(ctx, ids)
// }

// func (s *service) KillmailVictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error) {
// 	return s.KillmailRespository.KillmailVictimsByKillmailIDs(ctx, ids)
// }
