package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) VictimByKillmailID(ctx context.Context, id uint64, hash string) (*neo.KillmailVictim, error) {

	var victim = new(neo.KillmailVictim)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL_VICTIM, id, hash)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, &victim)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal killmail victim from redis")
		}
		return victim, nil
	}

	victim, err = s.victim.ByKillmailID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database for killmail")
	}

	byteSlc, err := json.Marshal(victim)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal killmail for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Hour).Result()

	return victim, errors.Wrap(err, "failed to cache killmail attackers in redis")

}

func (s *service) VictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error) {
	return s.victim.ByKillmailIDs(ctx, ids)
}
