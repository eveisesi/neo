package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) AttackersByKillmailID(ctx context.Context, id uint64, hash string) ([]*neo.KillmailAttacker, error) {

	var attackers = make([]*neo.KillmailAttacker, 0)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL_ATTACKERS, id, hash)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, &attackers)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal killmail from redis")
		}
		return attackers, nil
	}

	attackers, err = s.attackers.ByKillmailID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database for killmail")
	}

	byteSlc, err := json.Marshal(attackers)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal killmail for cache")
	}

	_, err = s.redis.WithContext(ctx).Set(key, byteSlc, time.Minute*60).Result()

	return attackers, errors.Wrap(err, "failed to cache killmail attackers in redis")

}

func (s *service) AttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error) {
	return s.attackers.ByKillmailIDs(ctx, ids)
}
