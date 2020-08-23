package resolvers

import (
	"context"
	"errors"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/models"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) Killmail(ctx context.Context, id int) (*neo.Killmail, error) {
	return r.Services.Killmail.Killmail(ctx, uint(id))
}

func (r *queryResolver) KillmailRecent(ctx context.Context, page *int) ([]*neo.Killmail, error) {

	if *page > 10 {
		*page = 10
	}
	return r.Services.Killmail.RecentKillmails(ctx, *page)

}

func (r *queryResolver) MvByEntityID(ctx context.Context, category *models.Category, entity *models.Entity, id *int, age *int, limit *int) ([]*neo.Killmail, error) {

	if *age > 14 {
		*age = 14
	}

	if *limit > 14 {
		*limit = 14
	}

	var mails []*neo.Killmail

	switch *category {
	case models.CategoryAll, models.CategoryKill:
		switch *entity {
		case models.EntityAll:
			mails, err = r.Services.Killmail.MostValuable(ctx, "none", 0, uint(*age), uint(*limit))
		case models.EntityCharacter:
			mails, err = r.Services.MostValuableKills(ctx, "character_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityCorporation:
			mails, err = r.Services.MostValuableKills(ctx, "corporation_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityAlliance:
			mails, err = r.Services.MostValuableKills(ctx, "alliance_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityShip:
			mails, err = r.Services.MostValuableKills(ctx, "ship_type_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityShipGroup:
			mails, err = r.Services.MostValuableKills(ctx, "ship_group_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntitySystem:
			mails, err = r.Services.MostValuable(ctx, "system_solar_id", uint(*id), uint(*age), uint(*limit))
		case models.EntityConstellation:
			mails, err = r.Services.MostValuable(ctx, "constellation_id", uint(*id), uint(*age), uint(*limit))
		case models.EntityRegion:
			mails, err = r.Services.MostValuable(ctx, "region_id", uint(*id), uint(*age), uint(*limit))
		default:
			return nil, errors.New("invalid entity")
		}
	case models.CategoryLose:
		switch *entity {
		case models.EntityAll:
			mails, err = r.Services.Killmail.MostValuable(ctx, "none", 0, uint(*age), uint(*limit))
		case models.EntityCharacter:
			mails, err = r.Services.MostValuableLosses(ctx, "character_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityCorporation:
			mails, err = r.Services.MostValuableLosses(ctx, "corporation_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityAlliance:
			mails, err = r.Services.MostValuableLosses(ctx, "alliance_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityShip:
			mails, err = r.Services.MostValuableLosses(ctx, "ship_type_id", uint64(*id), uint(*age), uint(*limit))
		case models.EntityShipGroup:
			mails, err = r.Services.MostValuableLosses(ctx, "ship_group_id", uint64(*id), uint(*age), uint(*limit))
		default:
			return nil, errors.New("invalid entity")
		}
	default:
		return nil, errors.New("invalid category specified")
	}

	return mails, err

}

func (r *queryResolver) KillmailsByEntityID(ctx context.Context, entity models.Entity, id int, page *int) ([]*neo.Killmail, error) {

	var mails []*neo.Killmail

	switch entity {
	case models.EntityAll:
		return nil, errors.New("All Type is not supported on this query")
	case models.EntityCharacter:
		mails, err = r.Services.KillmailsByCharacterID(ctx, uint64(id), uint(*page))
	case models.EntityCorporation:
		mails, err = r.Services.KillmailsByCorporationID(ctx, uint(id), uint(*page))
	case models.EntityAlliance:
		mails, err = r.Services.KillmailsByAllianceID(ctx, uint(id), uint(*page))
	case models.EntityShip:
		mails, err = r.Services.KillmailsByShipID(ctx, uint(id), uint(*page))
	case models.EntityShipGroup:
		mails, err = r.Services.KillmailsByShipGroupID(ctx, uint(id), uint(*page))
	case models.EntitySystem:
		mails, err = r.Services.KillmailsBySystemID(ctx, uint(id), uint(*page))
	case models.EntityConstellation:
		mails, err = r.Services.KillmailsByConstellationID(ctx, uint(id), uint(*page))
	case models.EntityRegion:
		mails, err = r.Services.KillmailsByRegionID(ctx, uint(id), uint(*page))
	default:
		return nil, errors.New("invalid entity")
	}

	return mails, err
}

func (r *Resolver) Killmail() service.KillmailResolver {
	return &killmailResolver{r}
}

type killmailResolver struct{ *Resolver }

func (r *killmailResolver) Attackers(ctx context.Context, obj *neo.Killmail) ([]*neo.KillmailAttacker, error) {
	return r.Dataloader(ctx).KillmailAttackersLoader.Load(obj.ID)
}

func (r *killmailResolver) Victim(ctx context.Context, obj *neo.Killmail) (*neo.KillmailVictim, error) {
	return r.Dataloader(ctx).KillmailVictimLoader.Load(obj.ID)
}

func (r *killmailResolver) System(ctx context.Context, obj *neo.Killmail) (*neo.SolarSystem, error) {
	return r.Dataloader(ctx).SolarSystemLoader.Load(obj.SolarSystemID)
}
