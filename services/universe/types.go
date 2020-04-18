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

var rtype = "type:%d"

func (s *service) Type(ctx context.Context, id uint64) (*neo.Type, error) {

	var invType = new(neo.Type)
	var key = fmt.Sprintf(rtype, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && !errors.Is(err, neo.ErrRedisNil) {
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	if err == nil {
		byteSlc, err := json.Marshal(invType)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal type for cache")
		}

		_, err = s.redis.Set(key, byteSlc, time.Minute*60).Result()

		return invType, errors.Wrap(err, "failed to cache type in redis")
	}

	// Type is not cached, the DB doesn't have this type, lets check ESI
	res, err := s.esi.GetUniverseTypesTypeID(id)
	if err != nil {
		return nil, errors.Wrap(err, "unable retrieve type from ESI")
	}

	// TODO: Check the res.Code. If not 200, do not proceed

	var ok bool
	if invType, ok = res.Data.(map[string]interface{})["type"].(*neo.Type); !ok {
		return nil, errors.Wrap(err, "unexpected results receive from ESI package")
	}

	var attributes []*neo.TypeAttribute
	if attributes, ok = res.Data.(map[string]interface{})["attributes"].([]*neo.TypeAttribute); !ok {
		return nil, errors.Wrap(err, "unexpected results receive from ESI package")
	}

	// ESI has the type. Lets insert it into the db, and cache it is redis
	err = s.UniverseRepository.CreateType(ctx, invType)
	if err != nil {
		return invType, errors.Wrap(err, "unable to insert type into db")
	}

	// ESI has the type attributes. Lets insert it into the db, and cache it is redis
	err = s.UniverseRepository.CreateTypeAttributes(ctx, attributes)
	if err != nil {
		return invType, errors.Wrap(err, "unable to insert type into db")
	}

	bType, err := json.Marshal(invType)
	if err != nil {
		return invType, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, bType, time.Minute*60).Result()

	return invType, errors.Wrap(err, "failed to cache solar type in redis")
}

func (s *service) TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.Type, error) {

	var types = make([]*neo.Type, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rtype, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && !errors.Is(err, neo.ErrRedisNil) {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var invType = new(neo.Type)
			err = json.Unmarshal(result, invType)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal invType bytes into struct")
			}

			types = append(types, invType)

		}
	}

	if len(ids) == len(types) {
		return types, nil
	}

	var missing []uint64
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

	dbTypes, err := s.UniverseRepository.TypesByTypeIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, invType := range dbTypes {
		key := fmt.Sprintf(rtype, invType.ID)

		bType, err := json.Marshal(invType)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal invType to slice of bytes")
		}

		_, err = s.redis.Set(key, bType, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache invType in redis")
		}

		types = append(types, invType)
	}

	return types, nil

}

var rattributes = "type:attributes:%d"

func (s *service) TypeAttributes(ctx context.Context, id uint64) ([]*neo.TypeAttribute, error) {

	var attributes = make([]*neo.TypeAttribute, 0)
	var key = fmt.Sprintf(rattributes, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && !errors.Is(err, neo.ErrRedisNil) {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, &attributes)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return attributes, nil
	}

	attributes, err = s.UniverseRepository.TypeAttributes(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database for type attributes")
	}

	byteSlc, err := json.Marshal(attributes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Minute*60).Result()

	return attributes, errors.Wrap(err, "failed to cache type in redis")
}

func (s *service) TypeAttributesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.TypeAttribute, error) {
	var final = make([]*neo.TypeAttribute, 0)

	var attributes = make(map[uint64][]*neo.TypeAttribute)

	for _, id := range ids {
		key := fmt.Sprintf(rattributes, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && !errors.Is(err, neo.ErrRedisNil) {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var typeAttributes = make([]*neo.TypeAttribute, 0)
			err = json.Unmarshal(result, &typeAttributes)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal invType bytes into struct")
			}

			attributes[id] = typeAttributes
		}
	}

	if len(ids) == len(attributes) {
		for _, typeAttributes := range attributes {
			final = append(final, typeAttributes...)
		}

		return final, nil
	}

	var missing []uint64
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

	dbAttributes, err := s.UniverseRepository.TypeAttributesByTypeIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing attributes")
	}

	for _, attribute := range dbAttributes {

		attributes[attribute.TypeID] = append(attributes[attribute.TypeID], attribute)

	}

	for typeID, typeAttributes := range attributes {
		key := fmt.Sprintf(rattributes, typeID)

		bAttributes, err := json.Marshal(typeAttributes)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal typeAttributes to slice of bytes")
		}

		_, err = s.redis.Set(key, bAttributes, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache invType in redis")
		}
	}

	for _, typeAttributes := range attributes {
		final = append(final, typeAttributes...)
	}

	return final, nil

}

var rcategory = "type:category:%d"

func (s *service) TypeCategory(ctx context.Context, id uint64) (*neo.TypeCategory, error) {

	var invCategory = new(neo.TypeCategory)
	var key = fmt.Sprintf(rcategory, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && !errors.Is(err, neo.ErrRedisNil) {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, invCategory)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return invCategory, nil
	}

	invCategory, err = s.UniverseRepository.TypeCategory(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlc, err := json.Marshal(invCategory)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Hour*24).Result()

	return invCategory, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint64) ([]*neo.TypeCategory, error) {

	var categories = make([]*neo.TypeCategory, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rcategory, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && !errors.Is(err, neo.ErrRedisNil) {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var invCategory = new(neo.TypeCategory)
			err = json.Unmarshal(result, invCategory)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal invCategory bytes into struct")
			}

			categories = append(categories, invCategory)

		}
	}

	if len(ids) == len(categories) {
		return categories, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, invCategory := range categories {
			if invCategory.ID == id {
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

	dbCategory, err := s.UniverseRepository.TypeCategoriesByCategoryIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, invCategory := range dbCategory {
		key := fmt.Sprintf(rcategory, invCategory.ID)

		bType, err := json.Marshal(invCategory)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal invCategory to slice of bytes")
		}

		_, err = s.redis.Set(key, bType, time.Hour*24).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache invCategory in redis")
		}

		categories = append(categories, invCategory)
	}

	return categories, nil

}

var rflag = "type:flag:%d"

func (s *service) TypeFlag(ctx context.Context, id uint64) (*neo.TypeFlag, error) {

	var invFlag = new(neo.TypeFlag)
	var key = fmt.Sprintf(rflag, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && !errors.Is(err, neo.ErrRedisNil) {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, invFlag)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return invFlag, nil
	}

	invFlag, err = s.UniverseRepository.TypeFlag(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlc, err := json.Marshal(invFlag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Hour*24).Result()

	return invFlag, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint64) ([]*neo.TypeFlag, error) {

	var flags = make([]*neo.TypeFlag, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rflag, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && !errors.Is(err, neo.ErrRedisNil) {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var invFlag = new(neo.TypeFlag)
			err = json.Unmarshal(result, invFlag)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal invFlag bytes into struct")
			}

			flags = append(flags, invFlag)

		}
	}

	if len(ids) == len(flags) {
		return flags, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, invFlag := range flags {
			if invFlag.ID == id {
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

	dbFlags, err := s.UniverseRepository.TypeFlagsByTypeFlagIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, invFlag := range dbFlags {
		key := fmt.Sprintf(rflag, invFlag.ID)

		bType, err := json.Marshal(invFlag)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal invFlag to slice of bytes")
		}

		_, err = s.redis.Set(key, bType, time.Hour*24).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache invFlag in redis")
		}

		flags = append(flags, invFlag)
	}

	return flags, nil

}

var rgroup = "type:group:%d"

func (s *service) TypeGroup(ctx context.Context, id uint64) (*neo.TypeGroup, error) {

	var invGroup = new(neo.TypeGroup)
	var key = fmt.Sprintf(rgroup, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && !errors.Is(err, neo.ErrRedisNil) {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, invGroup)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal type from redis")
		}
		return invGroup, nil
	}

	invGroup, err = s.UniverseRepository.TypeGroup(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for type")
	}

	byteSlc, err := json.Marshal(invGroup)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal type for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Hour*24).Result()

	return invGroup, errors.Wrap(err, "failed to cache category in redis")

}

func (s *service) TypeGroupsByGroupIDs(ctx context.Context, ids []uint64) ([]*neo.TypeGroup, error) {

	var groups = make([]*neo.TypeGroup, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rgroup, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && !errors.Is(err, neo.ErrRedisNil) {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var invGroup = new(neo.TypeGroup)
			err = json.Unmarshal(result, invGroup)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal invGroup bytes into struct")
			}

			groups = append(groups, invGroup)

		}
	}

	if len(ids) == len(groups) {
		return groups, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, invGroup := range groups {
			if invGroup.ID == id {
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

	dbGroups, err := s.UniverseRepository.TypeGroupsByGroupIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, invGroup := range dbGroups {
		key := fmt.Sprintf(rgroup, invGroup.ID)

		bType, err := json.Marshal(invGroup)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal invGroup to slice of bytes")
		}

		_, err = s.redis.Set(key, bType, time.Hour*24).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache invGroup in redis")
		}

		groups = append(groups, invGroup)
	}

	return groups, nil

}
