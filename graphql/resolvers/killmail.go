package resolvers

import (
	"context"
	"errors"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/models"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) Killmail(ctx context.Context, id int, hash string) (*neo.Killmail, error) {
	return r.Services.Killmail.Killmail(ctx, uint64(id), hash)
}

func (r *queryResolver) KillmailRecent(ctx context.Context, page *int) ([]*neo.Killmail, error) {

	newPage := 1
	if page != nil && *page > 0 {
		newPage = *page
	}

	return r.Services.Killmail.RecentKillmails(ctx, newPage)
}

func (r *queryResolver) MvkByEntityID(ctx context.Context, entity models.Entity, id *int, age *int, limit *int) ([]*neo.Killmail, error) {

	newID := uint64(0)
	if id != nil {
		newID = uint64(*id)
	}
	newAge := *age
	newLimit := *limit

	if newAge > 14 {
		newAge = 14
	}

	if newLimit > 14 {
		newLimit = 14
	}

	var killmails []*neo.Killmail

	switch entity {
	case models.EntityAll:
		killmails, err = r.Services.Killmail.MVKAll(ctx, newAge, newLimit)
	case models.EntityCharacter:
		killmails, err = r.Services.MVKByCharacterID(ctx, newID, newAge, newLimit)
	case models.EntityCorporation:
		killmails, err = r.Services.MVKByCorporationID(ctx, newID, newAge, newLimit)
	case models.EntityAlliance:
		killmails, err = r.Services.MVKByAllianceID(ctx, newID, newAge, newLimit)
	case models.EntityShip:
		killmails, err = r.Services.MVKByShipID(ctx, newID, newAge, newLimit)
	default:
		return nil, errors.New("invalid entity")
	}

	return killmails, err

}

func (r *queryResolver) KillmailsByEntityID(ctx context.Context, entity models.Entity, id int, page *int) ([]*neo.Killmail, error) {

	if *page > 0 {
		*page = *page - 1
		if *page > 10 {
			*page = 10
		}
	} else if *page < 0 {
		*page = 0
	}

	var killmails []*neo.Killmail

	switch entity {
	case models.EntityAll:
		return nil, errors.New("All Type is not supported on this query")
	case models.EntityCharacter:
		killmails, err = r.Services.KillmailsByCharacterID(ctx, uint64(id), *page)
	case models.EntityCorporation:
		killmails, err = r.Services.KillmailsByCorporationID(ctx, uint64(id), *page)
	case models.EntityAlliance:
		killmails, err = r.Services.KillmailsByAllianceID(ctx, uint64(id), *page)
	case models.EntityShip:
		return nil, errors.New("comming soom(tm)")
	default:
		return nil, errors.New("invalid entity")
	}

	return killmails, err
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
