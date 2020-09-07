package corporation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) Corporation(ctx context.Context, id uint) (*neo.Corporation, error) {
	var corporation = new(neo.Corporation)
	var key = fmt.Sprintf(neo.REDIS_CORPORATION, id)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for corporation")
	}

	if err == nil {
		bSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation for cache")
		}

		_, err = s.redis.WithContext(ctx).Set(key, bSlice, time.Minute*60).Result()

		return corporation, errors.Wrap(err, "failed to cache corporation in redis")
	}

	// Corporation is not cached, the DB doesn't have this corporation, lets check ESI
	corporation, m := s.esi.GetCorporationsCorporationID(ctx, id, "")
	if m.IsErr() {
		return nil, m.Msg
	}

	// ESI has the corporation. Lets insert it into the db, and cache it is redis
	err = s.CorporationRespository.CreateCorporation(ctx, corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to insert corporation into db")
	}

	byteSlice, err := json.Marshal(corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to marshal corporation for cache")
	}

	_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()

	return corporation, errors.Wrap(err, "failed to cache solar corporation in redis")
}

func (s *service) CorporationsByCorporationIDs(ctx context.Context, ids []uint) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_CORPORATION, id)
		result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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

	var missing []neo.ModValue
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

	dbTypes, err := s.Corporations(ctx, neo.In{Column: "id", Values: missing})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, corporation := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_CORPORATION, corporation.ID)

		byteSlice, err := json.Marshal(corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal corporation to slice of bytes")
		}

		_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache corporation in redis")
		}

		corporations = append(corporations, corporation)
	}

	return corporations, nil

}

func (s *service) UpdateExpired(ctx context.Context) {

	for {
		txn := s.newrelic.StartTransaction("updateExpiredCorporations")
		ctx = newrelic.NewContext(ctx, txn)
		expired, err := s.Expired(ctx)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			s.logger.WithError(err).Error("Failed to fetch expired corporations")
			return
		}

		if len(expired) == 0 {
			s.logger.Info("no expired corporations found")
			time.Sleep(time.Minute * 5)
			continue
		}

		s.logger.WithField("count", len(expired)).Info("updating expired corporations")

		for _, corporation := range expired {
			entry := s.logger.WithContext(ctx).WithField("corporationID", corporation.ID)
			s.tracker.GateKeeper(ctx)
			seg := txn.StartSegment("handleCorporation")
			seg.AddAttribute("id", corporation.ID)
			newCorporation, m := s.esi.GetCorporationsCorporationID(ctx, corporation.ID, corporation.Etag)
			if m.IsErr() {
				txn.NoticeError(m.Msg)
				entry.WithError(m.Msg).Error("failed to fetch corporation from esi")
				continue
			}

			switch m.Code {
			case http.StatusInternalServerError, http.StatusBadRequest, http.StatusNotFound:
				corporation.CachedUntil = time.Now().Add(time.Minute * 2).Unix()
				corporation.UpdateError++

				err = s.UpdateCorporation(ctx, corporation.ID, corporation)
			case http.StatusNotModified:

				if corporation.NotModifiedCount >= 2 && corporation.UpdatePriority < 2 {
					corporation.NotModifiedCount = 0
					corporation.UpdatePriority++
				} else {
					corporation.NotModifiedCount++
				}

				corporation.UpdateError = 0
				corporation.CachedUntil = time.Unix(newCorporation.CachedUntil, 0).AddDate(0, 0, int(corporation.UpdatePriority)).Unix()
				corporation.Etag = newCorporation.Etag

				err = s.UpdateCorporation(ctx, corporation.ID, corporation)
			case http.StatusOK:
				err = s.UpdateCorporation(ctx, corporation.ID, newCorporation)
			default:
				entry.WithField("status_code", m.Code).Error("unaccounted for status code received from esi service")
			}

			if err != nil {
				txn.NoticeError(err)
				entry.WithError(err).Error("failed to update corporation")
				continue
			}

			entry.WithField("status_code", m.Code).Debug("corporation successfully updated")
			seg.End()
			time.Sleep(time.Millisecond * 50)
		}
		s.logger.WithContext(ctx).WithField("count", len(expired)).Info("corporations successfully updated")
		txn.End()
		time.Sleep(time.Second)

	}

}
