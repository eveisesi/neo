package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/tools"
	"github.com/pkg/errors"
	"github.com/sirkon/go-format"
)

func (s *service) Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL, id, hash)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, killmail)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal killmail from redis")
		}
		return killmail, nil
	}

	killmail, err = s.killmails.Killmail(ctx, id, hash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for killmail")
	}

	byteSlc, err := json.Marshal(killmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal killmail for cache")
	}

	_, err = s.redis.Set(key, byteSlc, time.Minute*60).Result()

	return killmail, errors.Wrap(err, "failed to cache killmail in redis")

}

func (s *service) RecentKillmails(ctx context.Context, page int) ([]*neo.Killmail, error) {
	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.Recent(ctx, neo.KILLMAILS_PER_PAGE, offset)
}

func (s *service) KillmailsByCharacterID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "characters",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByCharacterID(ctx, id)
	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "characters",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page >= len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByCorporationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "corporations",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByCorporationID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "corporations",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByAllianceID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "alliances",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByAllianceID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "alliances",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByShipID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "ships",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByShipID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "ships",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByShipGroupID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	allowed := tools.IsGroupAllowed(id)
	if !allowed {
		return nil, errors.New("invalid group id. Only published group ids are allowed")
	}

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "shipGroup",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByShipGroupID(ctx, id)
	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "shipGroup",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsBySystemID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "systems",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.BySystemID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "systems",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByConstellationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "constellation",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByConstellationID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "constellation",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}

func (s *service) KillmailsByRegionID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type": "region",
		"id":   id,
		"page": page,
	})

	results, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	killmails, err = s.killmails.ByRegionID(ctx, id)

	if err != nil {
		return nil, err
	}

	kmChunk := ChunkSliceKillmails(killmails, 50)

	for i, chunk := range kmChunk {

		var innerKey = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
			"type": "region",
			"id":   id,
			"page": i,
		})

		bSlice, err := json.Marshal(chunk)
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}
