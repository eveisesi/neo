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

var rregion = "region:%d"

func (s *service) Region(ctx context.Context, id uint64) (*neo.Region, error) {

	var region = new(neo.Region)
	var key = fmt.Sprintf(rregion, id)

	result, err := s.redis.Get(key).Bytes()
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlice, err := json.Marshal(region)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Hour*24).Result()

	return region, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) RegionsByRegionIDs(ctx context.Context, ids []uint64) ([]*neo.Region, error) {

	var regions = make([]*neo.Region, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rregion, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var region = new(neo.Region)
			err = json.Unmarshal(result, region)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal region bytes into struct")
			}

			regions = append(regions, region)

		}
	}

	if len(ids) == len(regions) {
		return regions, nil
	}

	var missing []uint64
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

	dbRegions, err := s.UniverseRepository.RegionsByRegionIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, region := range dbRegions {
		key := fmt.Sprintf(rregion, region.ID)

		byteSlice, err := json.Marshal(region)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal region to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Hour*24).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache region in redis")
		}

		regions = append(regions, region)
	}

	return regions, nil

}
