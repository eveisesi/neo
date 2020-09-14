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

	result, err := s.redis.Get(key).Bytes()
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

	_, err = s.redis.Set(key, byteSlice, time.Hour*24).Result()

	return constellation, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) ConstellationsByConstellationIDs(ctx context.Context, ids []uint) ([]*neo.Constellation, error) {

	var constellations = make([]*neo.Constellation, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_CONSTELLATION, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var constellation = new(neo.Constellation)
			err = json.Unmarshal(result, constellation)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal constellation bytes into struct")
			}

			constellations = append(constellations, constellation)

		}
	}

	if len(ids) == len(constellations) {
		return constellations, nil
	}

	var missing []neo.ModValue
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

	for _, constellation := range dbConstellations {
		key := fmt.Sprintf(neo.REDIS_CONSTELLATION, constellation.ID)

		byteSlice, err := json.Marshal(constellation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal constellation to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Hour*24).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache constellation in redis")
		}

		constellations = append(constellations, constellation)
	}

	return constellations, nil

}
