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

func (s *service) BlueprintMaterials(ctx context.Context, id uint) ([]*neo.BlueprintMaterial, error) {

	var materials = make([]*neo.BlueprintMaterial, 0)
	var key = fmt.Sprintf(neo.REDIS_BLUEPRINT_MATERIALS, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, &materials)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return materials, nil
	}

	materials, err = s.BlueprintRepository.BlueprintMaterials(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database for type materials")
	}

	byteSlice, err := json.Marshal(materials)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour).Result()

	return materials, errors.Wrap(err, "failed to cache type in redis")
}

func (s *service) BlueprintProduct(ctx context.Context, id uint) (*neo.BlueprintProduct, error) {

	var product = new(neo.BlueprintProduct)
	var key = fmt.Sprintf(neo.REDIS_BLUEPRINT_PRODUCT, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, product)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal prodict from redis")
		}
		return product, nil
	}

	product, err = s.BlueprintRepository.BlueprintProduct(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for prodict")
	}

	byteSlice, err := json.Marshal(product)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal prodict for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return product, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) BlueprintProductByProductTypeID(ctx context.Context, id uint) (*neo.BlueprintProduct, error) {
	var product = new(neo.BlueprintProduct)
	var key = fmt.Sprintf(neo.REDIS_BLUEPRINT_PRODUCTTYPEID, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, product)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal prodict from redis")
		}
		return product, nil
	}

	product, err = s.BlueprintRepository.BlueprintProductByProductTypeID(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for prodict")
	}

	byteSlice, err := json.Marshal(product)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal prodict for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return product, errors.Wrap(err, "failed to cache category in redis")
}
