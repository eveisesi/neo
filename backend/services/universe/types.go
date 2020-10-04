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

func (s *service) Type(ctx context.Context, id uint) (*neo.Type, error) {

	var invType = new(neo.Type)
	var key = fmt.Sprintf(neo.REDIS_TYPE, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, invType)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return invType, nil
	}

	invType, err = s.UniverseRepository.Type(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	if err == nil {
		byteSlc, err := json.Marshal(invType)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal type for cache")
		}

		_, err = s.redis.Set(ctx, key, byteSlc, time.Hour).Result()

		return invType, errors.Wrap(err, "failed to cache type in redis")
	}

	// Type is not cached, the DB doesn't have this type, lets check ESI
	invType, attributes, m := s.esi.GetUniverseTypesTypeID(ctx, id)
	if m.IsErr() {
		return nil, m.Msg
	}

	// ESI has the type. Lets insert it into the db, and cache it is redis
	err = s.UniverseRepository.CreateType(ctx, invType)
	if err != nil {
		return invType, errors.Wrap(err, "unable to insert type into db")
	}

	if len(attributes) > 0 {
		// ESI has the type attributes. Lets insert it into the db, and cache it is redis
		err = s.UniverseRepository.CreateTypeAttributes(ctx, attributes)
		if err != nil {
			return invType, errors.Wrap(err, "unable to insert type into db")
		}
	}

	byteSlice, err := json.Marshal(invType)
	if err != nil {
		return invType, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour).Result()

	return invType, errors.Wrap(err, "failed to cache solar type in redis")
}

func (s *service) TypesByTypeIDs(ctx context.Context, ids []uint) ([]*neo.Type, error) {

	var types = make([]*neo.Type, 0)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_TYPE, id)
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
				var invType = new(neo.Type)
				err = json.Unmarshal([]byte(result), invType)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal invType bytes into struct")
				}

				types = append(types, invType)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(types) {
		return types, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, invType := range types {
			if invType.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return types, nil
	}

	dbTypes, err := s.UniverseRepository.Types(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, invType := range dbTypes {
		types = append(types, invType)

		key := fmt.Sprintf(neo.REDIS_TYPE, invType.ID)

		byteSlice, err := json.Marshal(invType)
		if err != nil {
			s.logger.WithError(err).WithField("id", invType.ID).Error("unable to marshal type to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache types in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return types, nil

}

func (s *service) TypeAttributes(ctx context.Context, id uint) ([]*neo.TypeAttribute, error) {

	var attributes = make([]*neo.TypeAttribute, 0)
	var key = fmt.Sprintf(neo.REDIS_TYPE_ATTRIBUTES, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, &attributes)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return attributes, nil
	}

	attributes, err = s.UniverseRepository.TypeAttributes(ctx, neo.NewEqualOperator("typeID", id))
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database for type attributes")
	}

	byteSlice, err := json.Marshal(attributes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour).Result()

	return attributes, errors.Wrap(err, "failed to cache type in redis")
}

func (s *service) TypeAttributesByTypeIDs(ctx context.Context, ids []uint) ([]*neo.TypeAttribute, error) {
	var final = make([]*neo.TypeAttribute, 0)

	var attributes = make(map[uint][]*neo.TypeAttribute)

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_TYPE_ATTRIBUTES, id)
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
				var typeAttributes = make([]*neo.TypeAttribute, 0)
				err = json.Unmarshal([]byte(result), &typeAttributes)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal typeAttributes bytes into struct")
				}
				// i here is the same i from the ids slice, since that i was used to construct the redis kye
				// so this should be fine since go-redis will
				// return a slice of something, be it the value we requested or nil, so the index i
				// should never be missing from the results
				attributes[ids[i]] = typeAttributes
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(attributes) {
		for _, typeAttributes := range attributes {
			final = append(final, typeAttributes...)
		}

		return final, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		if _, ok := attributes[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		for _, typeAttributes := range attributes {
			final = append(final, typeAttributes...)
		}

		return final, nil
	}

	dbAttributes, err := s.UniverseRepository.TypeAttributes(ctx, neo.NewInOperator("typeID", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing attributes")
	}

	for _, attribute := range dbAttributes {

		attributes[attribute.TypeID] = append(attributes[attribute.TypeID], attribute)

	}

	keyMap := make(map[string]interface{})

	for typeID, typeAttributes := range attributes {
		key := fmt.Sprintf(neo.REDIS_TYPE_ATTRIBUTES, typeID)

		byteSlice, err := json.Marshal(typeAttributes)
		if err != nil {
			s.logger.WithError(err).WithField("typeID", typeID).Error("unable to marshal typeAttributes for type to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache typeAttributes in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	for _, typeAttributes := range attributes {
		final = append(final, typeAttributes...)
	}

	return final, nil

}

func (s *service) TypeCategory(ctx context.Context, id uint) (*neo.TypeCategory, error) {

	var category = new(neo.TypeCategory)
	var key = fmt.Sprintf(neo.REDIS_TYPE_CATEGORY, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, category)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return category, nil
	}

	category, err = s.UniverseRepository.TypeCategory(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlice, err := json.Marshal(category)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return category, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint) ([]*neo.TypeCategory, error) {

	var categories = make([]*neo.TypeCategory, 0)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_TYPE_CATEGORY, id)
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
				var category = new(neo.TypeCategory)
				err = json.Unmarshal([]byte(result), category)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal category bytes into struct")
				}

				categories = append(categories, category)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(categories) {
		return categories, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, category := range categories {
			if category.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return categories, nil
	}

	dbCategory, err := s.UniverseRepository.TypeCategories(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, category := range dbCategory {

		categories = append(categories, category)

		key := fmt.Sprintf(neo.REDIS_TYPE_CATEGORY, category.ID)

		byteSlice, err := json.Marshal(category)
		if err != nil {
			s.logger.WithError(err).WithField("id", category.ID).Error("unable to marshal category to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)
	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache categories in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return categories, nil

}

func (s *service) TypeFlag(ctx context.Context, id uint) (*neo.TypeFlag, error) {

	var flag = new(neo.TypeFlag)
	var key = fmt.Sprintf(neo.REDIS_TYPE_FLAG, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, flag)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal flag from redis")
		}
		return flag, nil
	}

	flag, err = s.UniverseRepository.TypeFlag(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for flag")
	}

	byteSlice, err := json.Marshal(flag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal flag for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return flag, errors.Wrap(err, "failed to cache flag in redis")

}

func (s *service) TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint) ([]*neo.TypeFlag, error) {

	var flags = make([]*neo.TypeFlag, 0)

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_TYPE_FLAG, id)
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
				var flag = new(neo.TypeFlag)
				err = json.Unmarshal([]byte(result), flag)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal flag bytes into struct")
				}

				flags = append(flags, flag)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(flags) {
		return flags, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, flag := range flags {
			if flag.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return flags, nil
	}

	dbFlags, err := s.UniverseRepository.TypeFlags(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, flag := range dbFlags {

		flags = append(flags, flag)

		key := fmt.Sprintf(neo.REDIS_TYPE_FLAG, flag.ID)

		byteSlice, err := json.Marshal(flag)
		if err != nil {
			s.logger.WithError(err).WithField("id", flag.ID).Error("unable to marshal flag to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache flags in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return flags, nil

}

func (s *service) TypeGroup(ctx context.Context, id uint) (*neo.TypeGroup, error) {

	var group = new(neo.TypeGroup)
	var key = fmt.Sprintf(neo.REDIS_TYPE_GROUP, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, group)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return group, nil
	}

	group, err = s.UniverseRepository.TypeGroup(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlice, err := json.Marshal(group)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour*24).Result()

	return group, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) TypeGroupsByGroupIDs(ctx context.Context, ids []uint) ([]*neo.TypeGroup, error) {

	var groups = make([]*neo.TypeGroup, 0)

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_TYPE_GROUP, id)
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
				var group = new(neo.TypeGroup)
				err = json.Unmarshal([]byte(result), group)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal group bytes into struct")
				}

				groups = append(groups, group)
			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(groups) {
		return groups, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, group := range groups {
			if group.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return groups, nil
	}

	dbGroups, err := s.UniverseRepository.TypeGroups(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, group := range dbGroups {

		groups = append(groups, group)

		key := fmt.Sprintf(neo.REDIS_TYPE_GROUP, group.ID)

		byteSlice, err := json.Marshal(group)
		if err != nil {
			s.logger.WithError(err).WithField("id", group.ID).Error("unable to marshal group to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache groups in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return groups, nil

}
