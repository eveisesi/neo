package universe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) SolarSystem(ctx context.Context, id uint) (*neo.SolarSystem, error) {

	var system = new(neo.SolarSystem)
	var key = fmt.Sprintf(neo.REDIS_SYSTEM, id)

	result, err := s.redis.Get(ctx, key).Bytes()
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
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for solar system")
	}

	if err == nil {
		bSystem, err := json.Marshal(system)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal system for cache")
		}

		_, err = s.redis.Set(ctx, key, bSystem, time.Hour).Result()

		return system, errors.Wrap(err, "failed to cache solar system in redis")
	}

	// System is not cached, the DB doesn't have this system, lets check ESI
	system, m := s.esi.GetUniverseSystemsSystemID(ctx, id)
	if m.IsErr() {
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

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour).Result()

	return system, errors.Wrap(err, "failed to cache solar system in redis")
}

func (s *service) SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint) ([]*neo.SolarSystem, error) {

	var systems = make([]*neo.SolarSystem, 0)

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_SYSTEM, id)
	}

	results, err := s.redis.MGet(ctx, keys...).Result()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, errors.Wrap(err, "encountered error querying redis")
	}

	for i, resultInt := range results {
		if resultInt == nil {
			continue
		}

		switch result := resultInt.(type) {
		case string:
			if len(result) > 0 {
				var system = new(neo.SolarSystem)
				err = json.Unmarshal([]byte(result), system)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal system bytes into struct")
				}

				systems = append(systems, system)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(systems) {
		return systems, nil
	}

	var missing []neo.OpValue
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

	dbSystems, err := s.UniverseRepository.SolarSystems(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing solar system ids")
	}

	keyMap := make(map[string]interface{})

	for _, system := range dbSystems {
		systems = append(systems, system)

		key := fmt.Sprintf(neo.REDIS_SYSTEM, system.ID)

		byteSlice, err := json.Marshal(system)
		if err != nil {
			s.logger.WithError(err).WithField("id", system.ID).Error("unable to marshal system to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache systems in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return systems, nil

}
