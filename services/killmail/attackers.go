package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) AttackersByKillmailID(ctx context.Context, id uint) ([]*neo.KillmailAttacker, error) {

	var attackers = make([]*neo.KillmailAttacker, 0)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL_ATTACKERS, id)

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

func (s *service) AttackersByKillmailIDs(ctx context.Context, ids []uint) ([]*neo.KillmailAttacker, error) {

	var attackers = make([]*neo.KillmailAttacker, 0)
	var missing = make([]uint, 0)
	for _, id := range ids {

		key := fmt.Sprintf(neo.REDIS_KILLMAIL_ATTACKERS, id)
		result, err := s.redis.WithContext(ctx).Get(key).Bytes()
		if err != nil {
			missing = append(missing, id)
			continue
		}

		if len(result) > 0 {

			innerAttackers := make([]*neo.KillmailAttacker, 0)
			err = json.Unmarshal(result, &innerAttackers)
			if err != nil {
				missing = append(missing, id)
				continue
			}

			attackers = append(attackers, innerAttackers...)
			continue

		}

		missing = append(missing, id)
	}

	if len(missing) == 0 {
		return attackers, nil
	}

	missingAttackersByKillmailIDs, err := s.attackers.ByKillmailIDs(ctx, missing)
	if err != nil {
		return nil, err
	}

	if len(missingAttackersByKillmailIDs) == 0 {
		return attackers, nil
	}

	attackersByKillmailID := make(map[uint][]*neo.KillmailAttacker)
	for _, attacker := range missingAttackersByKillmailIDs {
		attackersByKillmailID[attacker.KillmailID] = append(attackersByKillmailID[attacker.KillmailID], attacker)
	}

	for i, v := range attackersByKillmailID {
		data, err := json.Marshal(v)
		if err != nil {
			continue
		}

		key := fmt.Sprintf(neo.REDIS_KILLMAIL_ATTACKERS, i)

		_ = s.redis.WithContext(ctx).Set(key, data, time.Minute*120)

	}

	attackers = append(attackers, missingAttackersByKillmailIDs...)

	return attackers, nil
}
