package killmail

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/sirkon/go-format"

	"github.com/eveisesi/neo"
)

var coreModFunc = func(age, limit uint) []neo.Modifier {
	return []neo.Modifier{
		neo.GreaterThanTime{Column: "killmail_time", Value: time.Now().AddDate(0, 0, 0-int(age))},
		neo.LimitModifier(int(limit)),
		neo.OrderModifier{Column: "total_value", Sort: neo.SortDesc},
	}
}

func (s *service) MostValuable(ctx context.Context, column string, id, age, limit uint) ([]*neo.Killmail, error) {
	var key = format.Formatm(neo.REDIS_MV_KILLMAILS, format.Values{
		"action": "all",
		"type":   strcase.ToLowerCamel(column),
		"id":     id,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	mods := coreModFunc(age, limit)
	if column != "none" && id > 0 {
		mods = append(mods, neo.EqualToUint{Column: column, Value: id})
	}

	killmails, err = s.killmails.Killmails(ctx, mods, []neo.Modifier{}, []neo.Modifier{})
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err

}

func (s *service) MostValuableKills(ctx context.Context, column string, id uint64, age, limit uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_MV_KILLMAILS, format.Values{
		"action": "kills",
		"type":   strcase.ToLowerCamel(column),
		"id":     id,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	mods := coreModFunc(age, limit)
	attMods := []neo.Modifier{
		neo.EqualToUint64{Column: column, Value: id},
	}

	killmails, err = s.killmails.Killmails(ctx, mods, []neo.Modifier{}, attMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err
}

func (s *service) MostValuableLosses(ctx context.Context, column string, id uint64, age, limit uint) ([]*neo.Killmail, error) {

	var key = format.Formatm(neo.REDIS_MV_KILLMAILS, format.Values{
		"action": "loses",
		"type":   strcase.ToLowerCamel(column),
		"id":     id,
	})

	killmails, err := s.KillmailsFromCache(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(killmails) > 0 {
		return killmails, nil
	}

	mods := coreModFunc(age, limit)
	vicMods := []neo.Modifier{
		neo.EqualToUint64{Column: column, Value: id},
	}

	killmails, err = s.killmails.Killmails(ctx, mods, []neo.Modifier{}, vicMods)
	if err != nil {
		return nil, err
	}

	err = s.CacheKillmailSlice(ctx, key, killmails)

	return killmails, err
}
