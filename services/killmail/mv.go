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
	gte := time.Date(now.Year(), now.Month(), now.Day()-age, now.Hour(), 0, 0, 0, time.UTC)
	lte := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 0, 0, 0, time.UTC)

	and := []neo.Modifier{
		neo.GreaterThanEqualTo{Column: "killmailTime", Value: gte},
		neo.LessThanEqualTo{Column: "killmailTime", Value: lte},
	}
	if column != "none" && id > 0 {
		and = append(and, neo.EqualTo{Column: column, Value: id})
	}

	mods := []neo.Modifier{
		neo.AndMod{
			Values: and,
		},
		neo.LimitModifier(limit),
		neo.OrderModifier{Column: "totalValue", Sort: neo.SortDesc},
	}

	modsMarshaled, err := json.Marshal(mods)
	if err != nil {
		fmt.Println(err)
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

	killmails, err = s.killmails.Killmails(ctx, mods...)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err
}
