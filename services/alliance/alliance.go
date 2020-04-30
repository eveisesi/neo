package alliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

func (s *service) Alliance(ctx context.Context, id uint64) (*neo.Alliance, error) {
	var alliance = new(neo.Alliance)
	var key = fmt.Sprintf(neo.REDIS_ALLIANCE, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal alliance from redis")
		}
		return alliance, nil
	}

	alliance, err = s.AllianceRespository.Alliance(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for alliance")
	}

	if err == nil {
		bSlice, err := json.Marshal(alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal alliance for cache")
		}

		_, err = s.redis.Set(key, bSlice, time.Minute*60).Result()

		return alliance, errors.Wrap(err, "failed to cache alliance in redis")
	}

	// Alliance is not cached, the DB doesn't have this alliance, lets check ESI
	alliance, m := s.esi.GetAlliancesAllianceID(id, null.NewString("", false))
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the alliance. Lets insert it into the db, and cache it is redis
	_, err = s.AllianceRespository.CreateAlliance(ctx, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance into db")
	}

	byteSlice, err := json.Marshal(alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to marshal alliance for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	return alliance, errors.Wrap(err, "failed to cache solar alliance in redis")
}

func (s *service) AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Alliance, error) {

	var alliances = make([]*neo.Alliance, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_ALLIANCE, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var alliance = new(neo.Alliance)
			err = json.Unmarshal(result, alliance)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal alliance bytes into struct")
			}

			alliances = append(alliances, alliance)

		}
	}

	if len(ids) == len(alliances) {
		return alliances, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, alliance := range alliances {
			if alliance.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return alliances, nil
	}

	dbTypes, err := s.AllianceRespository.AlliancesByAllianceIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, alliance := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_ALLIANCE, alliance.ID)

		byteSlice, err := json.Marshal(alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal alliance to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache alliance in redis")
		}

		alliances = append(alliances, alliance)
	}

	return alliances, nil

}
