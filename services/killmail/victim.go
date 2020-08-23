package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) VictimByKillmailID(ctx context.Context, id uint) (*neo.KillmailVictim, error) {

	var victim = new(neo.KillmailVictim)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL_VICTIM, id)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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

	_, err = s.redis.WithContext(ctx).Set(key, byteSlc, time.Hour).Result()

	return victim, errors.Wrap(err, "failed to cache killmail attackers in redis")

}

func (s *service) VictimsByKillmailIDs(ctx context.Context, ids []uint) ([]*neo.KillmailVictim, error) {

	var victims = make([]*neo.KillmailVictim, 0)
	var missing = make([]uint, 0)
	for _, id := range ids {

		key := fmt.Sprintf(neo.REDIS_KILLMAIL_VICTIM, id)
		result, err := s.redis.WithContext(ctx).Get(key).Bytes()
		if err != nil {
			missing = append(missing, id)
			continue
		}

		if len(result) > 0 {

			innerVictim := new(neo.KillmailVictim)
			err = json.Unmarshal(result, &innerVictim)
			if err != nil {
				missing = append(missing, id)
				continue
			}

			victims = append(victims, innerVictim)
			continue

		}

		missing = append(missing, id)

	}

	if len(missing) == 0 {
		return victims, nil
	}

	missingVictimsByKillmailIDs, err := s.victim.ByKillmailIDs(ctx, missing)
	if err != nil {
		return nil, err
	}

	if len(missingVictimsByKillmailIDs) == 0 {
		return victims, nil
	}

	victimByKillmailID := make(map[uint]*neo.KillmailVictim)
	for _, victim := range missingVictimsByKillmailIDs {
		victimByKillmailID[victim.KillmailID] = victim
	}

	for i, v := range victimByKillmailID {
		data, err := json.Marshal(v)
		if err != nil {
			continue
		}

		key := fmt.Sprintf(neo.REDIS_KILLMAIL_ATTACKERS, i)

		_ = s.redis.WithContext(ctx).Set(key, data, time.Minute*120)

	}

	victims = append(victims, missingVictimsByKillmailIDs...)

	return victims, nil
}
