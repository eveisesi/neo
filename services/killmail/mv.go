package killmail

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/sirkon/go-format"

	"github.com/eveisesi/neo"
)

func (s *service) MostValuable(ctx context.Context, column string, id uint64, age, limit int) ([]*neo.Killmail, error) {

	now := time.Now()
	and := []*neo.Operator{
		neo.NewGreaterThanEqualToOperator("killmailTime", time.Date(now.Year(), now.Month(), now.Day()-age, now.Hour(), 0, 0, 0, time.UTC)),
		neo.NewLessThanEqualToOperator("killmailTime", time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)),
	}
	if column != "none" && id > 0 {
		and = append(and, neo.NewEqualOperator(column, id))
	}

	operators := []*neo.Operator{
		neo.NewAndOperator(and...),
		neo.NewLimitOperator(int64(limit)),
		neo.NewOrderOperator("totalValue", neo.SortDesc),
	}

	modsMarshaled, err := json.Marshal(operators)
	if err != nil {
		return nil, err
	}

	var key = format.Formatm(neo.REDIS_MV_KILLMAILS, format.Values{
		"key":  strcase.ToLowerCamel(column),
		"id":   id,
		"mods": fmt.Sprintf("%x", sha256.Sum256(modsMarshaled)),
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	killmails, err = s.killmails.Killmails(ctx, operators...)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails, time.Minute)

	return killmails, err
}
