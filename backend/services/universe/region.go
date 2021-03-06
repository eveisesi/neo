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

func (s *service) Region(ctx context.Context, id uint) (*neo.Region, error) {

	var region = new(neo.Region)
	var key = fmt.Sprintf(neo.REDIS_REGION, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, region)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return region, nil
	}

	region, err = s.UniverseRepository.Region(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlice, err := json.Marshal(region)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return region, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) RegionsByRegionIDs(ctx context.Context, ids []uint) ([]*neo.Region, error) {

	var regions = make([]*neo.Region, 0)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_REGION, id)
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
				var region = new(neo.Region)
				err = json.Unmarshal([]byte(result), region)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal region bytes into struct")
				}

				regions = append(regions, region)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}
	if len(ids) == len(regions) {
		return regions, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, region := range regions {
			if region.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return regions, nil
	}

	dbRegions, err := s.UniverseRepository.Regions(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, region := range dbRegions {
		regions = append(regions, region)

		key := fmt.Sprintf(neo.REDIS_REGION, region.ID)

		byteSlice, err := json.Marshal(region)
		if err != nil {
			s.logger.WithError(err).WithField("id", region.ID).Error("unable to marshal region to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache regions in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return regions, nil

}
