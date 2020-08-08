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
	"github.com/sirupsen/logrus"
)

func (s *service) Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL, id, hash)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		s.logger.WithContext(ctx).WithError(err).WithField("key", key).Error("failed to fetch km from redis")
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

	_, err = s.redis.WithContext(ctx).Set(key, byteSlc, time.Minute*60).Result()

	return killmail, errors.Wrap(err, "failed to cache killmail in redis")

}

// FullKillmail assume that caller only needs ids. This function is not suitable if name resolution is needed
func (s *service) FullKillmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error) {

	var entry = s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   id,
		"hash": hash,
	})

	killmail, err := s.Killmail(ctx, id, hash)
	if err != nil {
		return nil, err
	}

	solarSystem, err := s.universe.SolarSystem(ctx, killmail.SolarSystemID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch solar system")
	}
	if err == nil {
		killmail.System = solarSystem
	}

	constellation, err := s.universe.Constellation(ctx, solarSystem.ConstellationID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch constellation")
	}
	if err == nil {
		solarSystem.Constellation = constellation
	}

	region, err := s.universe.Region(ctx, constellation.RegionID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch region")
	}
	if err == nil {
		constellation.Region = region
	}

	kmVictim, err := s.VictimByKillmailID(ctx, id, hash)
	if err != nil {
		entry.WithError(err).Error("Failed to retrieve killmail victim")
	}

	if kmVictim != nil {
		if kmVictim.CharacterID.Valid {
			character, err := s.character.Character(ctx, kmVictim.CharacterID.Uint64)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim character information")
			}
			if err == nil {
				kmVictim.Character = character
			}
		}
		if kmVictim.CorporationID.Valid {
			corporation, err := s.corporation.Corporation(ctx, kmVictim.CorporationID.Uint64)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim corporation information")
			}
			if err == nil {
				kmVictim.Corporation = corporation
			}
		}
		if kmVictim.AllianceID.Valid {
			alliance, err := s.alliance.Alliance(ctx, kmVictim.AllianceID.Uint64)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim alliance information")
			}
			if err == nil {
				kmVictim.Alliance = alliance
			}
		}
	}

	killmail.Victim = kmVictim

	ship, err := s.universe.Type(ctx, kmVictim.ShipTypeID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch ship")
	}
	if err == nil {
		killmail.Victim.Ship = ship
	}

	items, err := s.items.ByKillmailID(ctx, id)
	if err != nil {
		entry.WithError(err).Error("failed to fetch items")
	}
	if err == nil {
		kmVictim.Items = items
	}

	kmAttackers, err := s.AttackersByKillmailID(ctx, id, hash)
	if err != nil {
		entry.WithError(err).Error("failed to fetch km attackers")
	}

	killmail.Attackers = kmAttackers

	return killmail, nil
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.WithContext(ctx).Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.WithContext(ctx).Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.WithContext(ctx).Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

		_, err = s.redis.WithContext(ctx).Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
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

	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("unable to marshal chunk of killmails")
			continue
		}

		_, err = s.redis.Set(innerKey, bSlice, time.Minute*30).Result()
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithField("key", innerKey).Error("failed to cache killmail chunk in redis")
		}

	}

	if page > len(kmChunk) {
		return nil, nil
	}

	return kmChunk[page], nil

}
