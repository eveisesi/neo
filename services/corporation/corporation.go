package corporation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

func (s *service) Corporation(ctx context.Context, id uint64) (*neo.Corporation, error) {
	var corporation = new(neo.Corporation)
	var key = fmt.Sprintf(neo.REDIS_CORPORATION, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal corporation from redis")
		}
		return corporation, nil
	}

	corporation, err = s.CorporationRespository.Corporation(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for corporation")
	}

	if err == nil {
		bSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation for cache")
		}

		_, err = s.redis.Set(key, bSlice, time.Minute*60).Result()

		return corporation, errors.Wrap(err, "failed to cache corporation in redis")
	}

	// Corporation is not cached, the DB doesn't have this corporation, lets check ESI
	corporation, m := s.esi.GetCorporationsCorporationID(id, null.NewString("", false))
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the corporation. Lets insert it into the db, and cache it is redis
	_, err = s.CorporationRespository.CreateCorporation(ctx, corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to insert corporation into db")
	}

	byteSlice, err := json.Marshal(corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to marshal corporation for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	return corporation, errors.Wrap(err, "failed to cache solar corporation in redis")
}

func (s *service) CorporationsByCorporationIDs(ctx context.Context, ids []uint64) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_CORPORATION, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var corporation = new(neo.Corporation)
			err = json.Unmarshal(result, corporation)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal corporation bytes into struct")
			}

			corporations = append(corporations, corporation)

		}
	}

	if len(ids) == len(corporations) {
		return corporations, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, corporation := range corporations {
			if corporation.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return corporations, nil
	}

	dbTypes, err := s.CorporationRespository.CorporationsByCorporationIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, corporation := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_CORPORATION, corporation.ID)

		byteSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache corporation in redis")
		}

		corporations = append(corporations, corporation)
	}

	return corporations, nil

}

func (s *service) UpdateExpired(ctx context.Context) {

	for {
		expired, err := s.Expired(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("Failed to fetch expired corporations")
			return
		}

		if len(expired) == 0 {
			s.logger.Info("no expired corporations found")
			time.Sleep(time.Minute * 5)
			continue
		}

		for _, corporation := range expired {
			s.tracker.GateKeeper()
			newCorporation, m := s.esi.GetCorporationsCorporationID(corporation.ID, corporation.Etag)
			if m.IsError() {
				s.logger.WithError(err).WithField("corporation_id", corporation.ID).Error("failed to fetch corporation from esi")
				continue
			}

			switch m.Code {
			case http.StatusNotModified:

				corporation.NotModifiedCount++

				if corporation.NotModifiedCount >= 5 && corporation.UpdatePriority < 2 {
					corporation.NotModifiedCount = 0
					corporation.UpdatePriority++
				}

				corporation.CachedUntil = newCorporation.CachedUntil.AddDate(0, 0, int(corporation.UpdatePriority))
				corporation.Etag = newCorporation.Etag

				_, err = s.UpdateCorporation(ctx, corporation.ID, corporation)
			case http.StatusOK:
				_, err = s.UpdateCorporation(ctx, corporation.ID, newCorporation)
			default:
				s.logger.WithField("status_code", m.Code).Error("unaccounted for status code received from esi service")
			}

			if err != nil {
				s.logger.WithError(err).WithField("corporation_id", corporation.ID).Error("failed to update corporation")
				continue
			}

			s.logger.WithField("corporation_id", corporation.ID).WithField("status_code", m.Code).Info("corporation successfully updated")
		}
		time.Sleep(time.Second * 15)

	}

}
