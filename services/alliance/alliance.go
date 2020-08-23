package alliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

func (s *service) Alliance(ctx context.Context, id uint) (*neo.Alliance, error) {
	var alliance = new(neo.Alliance)
	var key = fmt.Sprintf(neo.REDIS_ALLIANCE, id)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for alliance")
	}

	if err == nil {
		bSlice, err := json.Marshal(alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal alliance for cache")
		}

		_, err = s.redis.WithContext(ctx).Set(key, bSlice, time.Minute*60).Result()

		return alliance, errors.Wrap(err, "failed to cache alliance in redis")
	}

	// Alliance is not cached, the DB doesn't have this alliance, lets check ESI
	alliance, m := s.esi.GetAlliancesAllianceID(ctx, id, null.NewString("", false))
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the alliance. Lets insert it into the db, and cache it is redis
	_, err = s.AllianceRespository.CreateAlliance(ctx, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance into db")
	}

	byteSlice, err := json.Marshal(alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to marshal alliance for cache")
	}

	_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()

	return alliance, errors.Wrap(err, "failed to cache solar alliance in redis")
}

func (s *service) AlliancesByAllianceIDs(ctx context.Context, ids []uint) ([]*neo.Alliance, error) {

	var alliances = make([]*neo.Alliance, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_ALLIANCE, id)
		result, err := s.redis.WithContext(ctx).Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var alliance = new(neo.Alliance)
			err = json.Unmarshal(result, alliance)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal alliance bytes into struct")
			}

			alliances = append(alliances, alliance)

		}
	}

	if len(ids) == len(alliances) {
		return alliances, nil
	}

	var missing []uint
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

	dbTypes, err := s.Alliances(ctx, neo.InUint{Column: "id", Value: missing})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, alliance := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_ALLIANCE, alliance.ID)

		byteSlice, err := json.Marshal(alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal alliance to slice of bytes")
		}

		_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache alliance in redis")
		}

		alliances = append(alliances, alliance)
	}

	return alliances, nil

}

func (s *service) UpdateExpired(ctx context.Context) {

	for {
		txn := s.newrelic.StartTransaction("updateExpiredAlliances")
		ctx = newrelic.NewContext(ctx, txn)

		expired, err := s.Expired(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			txn.NoticeError(err)
			s.logger.WithError(err).Error("Failed to fetch expired alliances")
			return
		}

		if len(expired) == 0 {
			s.logger.Info("no expired alliances found")
			time.Sleep(time.Minute * 5)
			continue
		}

		for _, alliance := range expired {
			seg := txn.StartSegment("handleAlliance")
			seg.AddAttribute("id", alliance.ID)
			s.tracker.GateKeeper(ctx)

			newAlliance, m := s.esi.GetAlliancesAllianceID(ctx, alliance.ID, alliance.Etag)
			if m.IsError() {
				txn.NoticeError(m.Msg)
				s.logger.WithError(m.Msg).WithField("alliance_id", alliance.ID).Error("failed to fetch alliance from esi")
				continue
			}

			switch m.Code {
			case http.StatusNotModified:

				alliance.NotModifiedCount++

				if alliance.NotModifiedCount >= 2 && alliance.UpdatePriority < 2 {
					alliance.NotModifiedCount = 0
					alliance.UpdatePriority++
				}

				alliance.CachedUntil = newAlliance.CachedUntil.Add(time.Hour*24).AddDate(0, 0, int(alliance.UpdatePriority))
				alliance.Etag = newAlliance.Etag

				_, err = s.UpdateAlliance(ctx, alliance.ID, alliance)
			case http.StatusOK:
				_, err = s.UpdateAlliance(ctx, alliance.ID, newAlliance)
			default:
				s.logger.WithField("status_code", m.Code).WithField("alliance_id", alliance.ID).Error("unaccounted for status code received from esi service")
			}

			if err != nil {
				txn.NoticeError(err)
				s.logger.WithError(err).WithField("alliance_id", alliance.ID).Error("failed to update alliance")
			}

			s.logger.WithField("alliance_id", alliance.ID).WithField("status_code", m.Code).Debug("alliance successfully updated")
			seg.End()

			time.Sleep(time.Millisecond * 25)
		}
		s.logger.WithField("count", len(expired)).Debug("alliances successfully updated")
		txn.End()
		time.Sleep(time.Second)

	}

}
