package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

var rkillmail = "killmail:%s:%s"

func (s *service) Killmail(ctx context.Context, id, hash string, ignoreCache, goRemote bool) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)
	var key = fmt.Sprintf(rkillmail, id, hash)

	if !ignoreCache {
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

		killmailID, _ := strconv.ParseUint(id, 10, 64)

		killmail, err = s.KillmailRespository.Killmail(ctx, killmailID, hash)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "unable to query database for killmail")
		}

		if err == nil {
			byteSlc, err := json.Marshal(killmail)
			if err != nil {
				return nil, errors.Wrap(err, "unable to marshal killmail for cache")
			}

			_, err = s.redis.Set(key, byteSlc, time.Minute*60).Result()

			return killmail, errors.Wrap(err, "failed to cache killmail in redis")
		}
	}

	if !goRemote {
		return killmail, errors.New("killmail not found")
	}

	// Type is not cached, the DB doesn't have this type, lets check ESI
	res, err := s.esi.GetKillmailsKillmailIDKillmailHash(id, hash)
	if err != nil {
		return nil, errors.Wrap(err, "unable retrieve type from ESI")
	}

	killmail = res.Data.(*neo.Killmail)

	return killmail, nil

	// // ESI has the type. Lets insert it into the db, and cache it is redis
	// err = s.KillmailRespository.CreateType(ctx, invType)
	// if err != nil {
	// 	return invType, errors.Wrap(err, "unable to insert type into db")
	// }

	// // ESI has the type attributes. Lets insert it into the db, and cache it is redis
	// err = s.UniverseRepository.CreateTypeAttributes(ctx, attributes)
	// if err != nil {
	// 	return invType, errors.Wrap(err, "unable to insert type into db")
	// }

	// byteSlice, err := json.Marshal(invType)
	// if err != nil {
	// 	return invType, errors.Wrap(err, "unable to marshal type for cache")
	// }

	// _, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	// return invType, errors.Wrap(err, "failed to cache solar type in redis")

}
