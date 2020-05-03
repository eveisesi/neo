package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
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

	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.ByCharacterID(ctx, id, neo.KILLMAILS_PER_PAGE, offset)

}

func (s *service) KillmailsByCorporationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.ByCorporationID(ctx, id, neo.KILLMAILS_PER_PAGE, offset)

}

func (s *service) KillmailsByAllianceID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.ByAllianceID(ctx, id, neo.KILLMAILS_PER_PAGE, offset)

}

func (s *service) KillmailsByShipID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error) {

	offset := (neo.KILLMAILS_PER_PAGE * page) - neo.KILLMAILS_PER_PAGE

	return s.killmails.ByShipID(ctx, id, neo.KILLMAILS_PER_PAGE, offset)

}
