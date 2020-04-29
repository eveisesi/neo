package corporation

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

var rcorporation = "corporation:%d"

func (s *service) Corporation(ctx context.Context, id uint64) (*neo.Corporation, error) {
	var corporation = new(neo.Corporation)
	var key = fmt.Sprintf(rcorporation, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal corporation from redis")
		}
		return corporation, nil
	}

	corporation, err = s.CorporationRespository.Corporation(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for corporation")
	}

	if err == nil {
		bSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation for cache")
		}

		_, err = s.redis.Set(key, bSlice, time.Minute*60).Result()

		return corporation, errors.Wrap(err, "failed to cache corporation in redis")
	}

	// Corporation is not cached, the DB doesn't have this corporation, lets check ESI
	corporation, m := s.esi.GetCorporationsCorporationID(id, null.NewString("", false))
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the corporation. Lets insert it into the db, and cache it is redis
	_, err = s.CorporationRespository.CreateCorporation(ctx, corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to insert corporation into db")
	}

	byteSlice, err := json.Marshal(corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to marshal corporation for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	return corporation, errors.Wrap(err, "failed to cache solar corporation in redis")
}

func (s *service) AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rcorporation, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var corporation = new(neo.Corporation)
			err = json.Unmarshal(result, corporation)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal corporation bytes into struct")
			}

			corporations = append(corporations, corporation)

		}
	}

	if len(ids) == len(corporations) {
		return corporations, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, corporation := range corporations {
			if corporation.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return corporations, nil
	}

	dbTypes, err := s.CorporationRespository.CorporationsByCorporationIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, corporation := range dbTypes {
		key := fmt.Sprintf(rcorporation, corporation.ID)

		byteSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache corporation in redis")
		}

		corporations = append(corporations, corporation)
	}

	return corporations, nil

}
