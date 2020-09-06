package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/tools"
	"github.com/pkg/errors"
	"github.com/sirkon/go-format"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) Killmail(ctx context.Context, id uint) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)
	var key = fmt.Sprintf(neo.REDIS_KILLMAIL, id)

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

	killmail, err = s.killmails.Killmail(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
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
func (s *service) FullKillmail(ctx context.Context, id uint, withNames bool) (*neo.Killmail, error) {

	var entry = s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id": id,
	})

	killmail, err := s.Killmail(ctx, id)
	if err != nil {
		return nil, err
	}

	if withNames {
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
	}

	kmVictim, err := s.VictimByKillmailID(ctx, id)
	if err != nil {
		entry.WithError(err).Error("Failed to retrieve killmail victim")
	}

	if kmVictim != nil && withNames {
		if kmVictim.CharacterID != nil {
			character, err := s.character.Character(ctx, *kmVictim.CharacterID)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim character information")
			}
			if err == nil {
				kmVictim.Character = character
			}
		}
		if kmVictim.CorporationID != nil {
			corporation, err := s.corporation.Corporation(ctx, *kmVictim.CorporationID)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim corporation information")
			}
			if err == nil {
				kmVictim.Corporation = corporation
			}
		}
		if kmVictim.AllianceID != nil {
			alliance, err := s.alliance.Alliance(ctx, *kmVictim.AllianceID)
			if err != nil {
				entry.WithError(err).Error("failed to fetch victim alliance information")
			}
			if err == nil {
				kmVictim.Alliance = alliance
			}
		}
	}

	killmail.Victim = kmVictim

	if withNames {
		ship, err := s.universe.Type(ctx, kmVictim.ShipTypeID)
		if err != nil {
			entry.WithError(err).Error("failed to fetch ship")
		}
		if err == nil {
			killmail.Victim.Ship = ship
		}
	}

	return killmail, nil
}

func (s *service) RecentKillmails(ctx context.Context, page int) ([]*neo.Killmail, error) {
	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.Recent(ctx, neo.KILLMAILS_PER_PAGE, offset)
}

func (s *service) KillmailsFromCache(ctx context.Context, key string) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	results, err := s.redis.WithContext(ctx).Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(results) > 0 {
		err = json.Unmarshal(results, &killmails)

		return killmails, errors.Wrap(err, "unable to unmarshal killmails from cache")
	}

	return nil, nil

}

func (s *service) CacheKillmailSlice(ctx context.Context, key string, killmails []*neo.Killmail) error {

	bSlice, err := json.Marshal(killmails)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithField("key", key).Error("unable to marshal killmails for cache")
		return err
	}

	_, err = s.redis.WithContext(ctx).Set(key, bSlice, time.Minute*30).Result()
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithField("key", key).Error("failed to cache killmail chunk in redis")
	}

	return err

}

func (s *service) KillmailsByCharacterID(ctx context.Context, id uint64, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "characters",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	mods := []neo.Modifier{
		neo.EqualTo{Column: "victim.characterID", Value: id},
		neo.EqualTo{Column: "attackers.characterID", Value: id},
	}

	if after > 0 {
		mods = append(mods, neo.LessThan{Column: "id", Value: after})
	}

	killmails, err = s.killmails.Killmails(ctx, mods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByCorporationID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "corporations",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}
	victimMods := []neo.Modifier{neo.EqualToUint{Column: "corporation_id", Value: id}}
	attMods := []neo.Modifier{neo.EqualToUint{Column: "corporation_id", Value: id}}

	killmails, err = s.killmails.Killmails(ctx, coreMods, victimMods, attMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByAllianceID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "alliances",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}
	victimMods := []neo.Modifier{neo.EqualToUint{Column: "alliance_id", Value: id}}
	attMods := []neo.Modifier{neo.EqualToUint{Column: "alliance_id", Value: id}}

	killmails, err = s.killmails.Killmails(ctx, coreMods, victimMods, attMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByShipID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "ships",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}
	victimMods := []neo.Modifier{neo.EqualToUint{Column: "ship_type_id", Value: id}}
	attMods := []neo.Modifier{neo.EqualToUint{Column: "ship_type_id", Value: id}}

	killmails, err = s.killmails.Killmails(ctx, coreMods, victimMods, attMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByShipGroupID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	allowed := tools.IsGroupAllowed(id)
	if !allowed {
		return nil, errors.New("invalid group id. Only published group ids are allowed")
	}

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "shipGroup",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}
	vicMods := []neo.Modifier{neo.EqualToUint{Column: "ship_group_id", Value: id}}
	attMods := []neo.Modifier{neo.EqualToUint{Column: "ship_group_id", Value: id}}

	killmails, err = s.killmails.Killmails(ctx, coreMods, vicMods, attMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsBySystemID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "systems",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{neo.EqualToUint{Column: "solar_system_id", Value: id}}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}

	killmails, err = s.killmails.Killmails(ctx, coreMods, []neo.Modifier{}, []neo.Modifier{})
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByConstellationID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "constellation",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{neo.EqualToUint{Column: "constellation_id", Value: id}}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}

	killmails, err = s.killmails.Killmails(ctx, coreMods, []neo.Modifier{}, []neo.Modifier{})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) KillmailsByRegionID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_KILLMAILS_BY_ENTITY, format.Values{
		"type":  "region",
		"id":    id,
		"after": after,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	coreMods := []neo.Modifier{neo.EqualToUint{Column: "region_id", Value: id}}
	if after > 0 {
		coreMods = append(coreMods, neo.LessThanUint{Column: "id", Value: after})
	}

	killmails, err = s.killmails.Killmails(ctx, coreMods, []neo.Modifier{}, []neo.Modifier{})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}
