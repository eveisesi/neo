package alliance

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

func (s *service) Alliance(ctx context.Context, id uint) (*neo.Alliance, error) {
	var alliance = new(neo.Alliance)
	var key = fmt.Sprintf(neo.REDIS_ALLIANCE, id)

	result, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal alliance from redis")
		}
		return alliance, nil
	}

	alliance, err = s.AllianceRespository.Alliance(ctx, id)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for alliance")
	}

	if err == nil {
		bSlice, err := json.Marshal(alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal alliance for cache")
		}

		_, err = s.redis.Set(ctx, key, bSlice, time.Hour).Result()

		return alliance, errors.Wrap(err, "failed to cache alliance in redis")
	}

	// Alliance is not cached, the DB doesn't have this alliance, lets check ESI
	alliance, m := s.esi.GetAlliancesAllianceID(ctx, id, "")
	if m.IsErr() {
		return nil, m.Msg
	}

	if m.Code == http.StatusUnprocessableEntity {
		return nil, errors.New("invalid alliance received from ESI, skipping create and cache")
	}

	// ESI has the alliance. Lets insert it into the db, and cache it is redis
	err = s.AllianceRespository.CreateAlliance(ctx, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance into db")
	}

	byteSlice, err := json.Marshal(alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to marshal alliance for cache")
	}

	_, err = s.redis.Set(ctx, key, byteSlice, time.Hour).Result()

	return alliance, errors.Wrap(err, "failed to cache solar alliance in redis")
}

func (s *service) AlliancesByAllianceIDs(ctx context.Context, ids []uint) ([]*neo.Alliance, error) {

	var alliances = make([]*neo.Alliance, 0)
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(neo.REDIS_ALLIANCE, id)
	}

	results, err := s.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, "encountered error querying redis")
	}

	for i, resultInt := range results {
		if resultInt == nil {
			continue
		}

		switch result := resultInt.(type) {
		case string:
			if len(result) > 0 {

				var alliance = new(neo.Alliance)
				err = json.Unmarshal([]byte(result), alliance)
				if err != nil {
					return nil, errors.Wrap(err, "unable to unmarshal alliance bytes into struct")
				}

				alliances = append(alliances, alliance)

			}
		default:
			panic(fmt.Sprintf("unexpected type received from redis. expected string, got %#T. redis key is %s", result, keys[i]))
		}
	}

	if len(ids) == len(alliances) {
		return alliances, nil
	}

	var missing []neo.OpValue
	for _, id := range ids {
		found := false
		for _, alliance := range alliances {
			if alliance.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return alliances, nil
	}

	dbTypes, err := s.Alliances(ctx, neo.NewInOperator("id", missing))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	keyMap := make(map[string]interface{})

	for _, alliance := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_ALLIANCE, alliance.ID)

		byteSlice, err := json.Marshal(alliance)
		if err != nil {
			s.logger.WithError(err).WithField("id", alliance.ID).Error("unable to marshal alliance to slice of bytes")
			continue
		}

		keyMap[key] = string(byteSlice)

		alliances = append(alliances, alliance)
	}

	_, err = s.redis.MSet(ctx, keyMap).Result()
	if err != nil {
		return nil, errors.Wrap(err, "unable to cache alliances in redis")
	}

	for i := range keyMap {
		s.redis.Expire(ctx, i, time.Hour)
	}

	return alliances, nil

}

func (s *service) UpdateExpired(ctx context.Context) {

	for {

		expired, err := s.Expired(ctx)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			s.logger.WithContext(ctx).WithError(err).Error("Failed to fetch expired alliances")
			return
		}

		if len(expired) == 0 {
			s.logger.WithContext(ctx).Info("no expired alliances found")
			time.Sleep(time.Minute * 5)
			continue
		}

		s.logger.WithField("count", len(expired)).Info("updating expired alliances")

		for _, alliance := range expired {
			s.tracker.Watchman(ctx)
			entry := s.logger.WithContext(ctx).WithField("allianceID", alliance.ID)

			txn := s.newrelic.StartTransaction("update-expired-alliance")
			ctx = newrelic.NewContext(ctx, txn)
			txn.AddAttribute("allianceID", alliance.ID)

			newAlliance, m := s.esi.GetAlliancesAllianceID(ctx, alliance.ID, alliance.Etag)
			if m.IsErr() {
				txn.NoticeError(m.Msg)
				txn.End()
				entry.WithError(m.Msg).Error("failed to fetch alliance from esi")
				continue
			}

			entry = entry.WithField("status_code", m.Code)
			txn.AddAttribute("status_code", m.Code)
			switch m.Code {
			case http.StatusInternalServerError, http.StatusBadRequest, http.StatusNotFound, http.StatusUnprocessableEntity:
				err = errors.New("bad status code received from ESI")
				txn.NoticeError(err)
				entry.WithError(err).Errorln()
				alliance.CachedUntil = time.Now().Add(time.Minute * 2).Unix()
				alliance.UpdateError++

				err = s.UpdateAlliance(ctx, alliance.ID, alliance)
			case http.StatusNotModified:

				if alliance.NotModifiedCount >= 2 && alliance.UpdatePriority < 2 {
					alliance.NotModifiedCount = 0
					alliance.UpdatePriority++
				} else {
					alliance.NotModifiedCount++
				}

				alliance.UpdateError = 0
				alliance.CachedUntil = time.Unix(newAlliance.CachedUntil, 0).AddDate(0, 0, int(alliance.UpdatePriority)).Unix()
				alliance.Etag = newAlliance.Etag

				err = s.UpdateAlliance(ctx, alliance.ID, alliance)
			case http.StatusOK:
				err = s.UpdateAlliance(ctx, alliance.ID, newAlliance)
			default:
				entry.WithField("status_code", m.Code).Error("unaccounted for status code received from esi service")
			}

			if err != nil {
				txn.NoticeError(err)
				entry.WithError(err).Error("failed to update alliance")
			}

			txn.End()

			time.Sleep(time.Millisecond * 50)
		}
		s.logger.WithField("count", len(expired)).Info("alliances successfully updated")
		time.Sleep(time.Second)

	}

}
