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

func (s *service) Constellation(ctx context.Context, id uint) (*neo.Constellation, error) {

	var constellation = new(neo.Constellation)
	var key = fmt.Sprintf(neo.REDIS_CONSTELLATION, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, constellation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return constellation, nil
	}

	constellation, err = s.UniverseRepository.Constellation(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlice, err := json.Marshal(constellation)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return constellation, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) ConstellationsByConstellationIDs(ctx context.Context, ids []uint) ([]*neo.Constellation, error) {

	var constellations = make([]*neo.Constellation, 0)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_CONSTELLATION, id)
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
				var constellation = new(neo.Constellation)
				err = json.Unmarshal([]byte(result), constellation)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal constellation bytes into struct")
				}

				constellations = append(constellations, constellation)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(constellations) {
		return constellations, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, constellation := range constellations {
			if constellation.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return constellations, nil
	}

	dbConstellations, err := s.UniverseRepository.Constellations(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, constellation := range dbConstellations {
		constellations = append(constellations, constellation)

		key := fmt.Sprintf(neo.REDIS_CONSTELLATION, constellation.ID)

		byteSlice, err := json.Marshal(constellation)
		if err != nil {
			s.logger.WithError(err).WithField("id", constellation.ID).Error("unable to marshal constellation to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache constellations in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return constellations, nil

}
