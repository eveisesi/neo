package universe

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) SolarSystem(ctx context.Context, id uint) (*neo.SolarSystem, error) {

	var system = new(neo.SolarSystem)
	var key = fmt.Sprintf(neo.REDIS_SYSTEM, id)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, system)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal system from redis")
		}
		return system, nil
	}

	system, err = s.UniverseRepository.SolarSystem(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for solar system")
	}

	if err == nil {
		bSystem, err := json.Marshal(system)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal system for cache")
		}

		_, err = s.redis.Set(key, bSystem, time.Minute*60).Result()

		return system, errors.Wrap(err, "failed to cache solar system in redis")
	}

	// System is not cached, the DB doesn't have this system, lets check ESI
	system, m := s.esi.GetUniverseSystemsSystemID(ctx, id)
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the system. Lets insert it into the db, and cache it is redis
	err = s.UniverseRepository.CreateSolarSystem(ctx, system)
	if err != nil {
		return system, errors.Wrap(err, "unable to insert system into db")
	}

	byteSlice, err := json.Marshal(system)
	if err != nil {
		return system, errors.Wrap(err, "unable to marshal system for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	return system, errors.Wrap(err, "failed to cache solar system in redis")
}

func (s *service) SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint) ([]*neo.SolarSystem, error) {

	var systems = make([]*neo.SolarSystem, 0)
	for _, v := range ids {
		key := fmt.Sprintf(neo.REDIS_SYSTEM, v)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var system = new(neo.SolarSystem)
			err = json.Unmarshal(result, system)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal system bytes into struct")
			}

			systems = append(systems, system)

		}

	}

	if len(ids) == len(systems) {
		return systems, nil
	}

	var missing []uint
	for _, id := range ids {
		found := false
		for _, system := range systems {
			if system.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return systems, nil
	}

	dbSystems, err := s.UniverseRepository.SolarSystemsBySolarSystemIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing solar system ids")
	}

	for _, system := range dbSystems {
		key := fmt.Sprintf(neo.REDIS_SYSTEM, system.ID)

		byteSlice, err := json.Marshal(system)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal system to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache solar system in redis")
		}

		systems = append(systems, system)
	}

	return systems, nil

}
