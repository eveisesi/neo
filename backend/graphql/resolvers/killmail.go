package resolvers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/models"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) Killmail(ctx context.Context, id int) (*neo.Killmail, error) {
	return r.Services.Killmail.Killmail(ctx, uint(id))
}

func (r *queryResolver) KillmailRecent(ctx context.Context, page *int) ([]*neo.Killmail, error) {

	r.Logger.Info("Start KillmailRecent")
	defer r.Logger.Info("End KillmailRecent")
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
			mails, err = r.Services.MostValuable(ctx, "none", 0, *age, *limit)
		case models.EntityCharacter:
			mails, err = r.Services.MostValuable(ctx, "attackers.characterID", uint64(*id), *age, *limit)
		case models.EntityCorporation:
			mails, err = r.Services.MostValuable(ctx, "attackers.corporationID", uint64(*id), *age, *limit)
		case models.EntityAlliance:
			mails, err = r.Services.MostValuable(ctx, "attackers.allianceID", uint64(*id), *age, *limit)
		case models.EntityShip:
			mails, err = r.Services.MostValuable(ctx, "attackers.shipTypeID", uint64(*id), *age, *limit)
		case models.EntityShipGroup:
			mails, err = r.Services.MostValuable(ctx, "attackers.shipGroupID", uint64(*id), *age, *limit)
		case models.EntitySystem:
			mails, err = r.Services.MostValuable(ctx, "solarSystemID", uint64(*id), *age, *limit)
		case models.EntityConstellation:
			mails, err = r.Services.MostValuable(ctx, "constellationID", uint64(*id), *age, *limit)
		case models.EntityRegion:
			mails, err = r.Services.MostValuable(ctx, "regionID", uint64(*id), *age, *limit)
		default:
			return nil, errors.New("invalid entity")
		}
	case models.CategoryLose:
		switch *entity {
		case models.EntityAll:
			mails, err = r.Services.Killmail.MostValuable(ctx, "none", 0, *age, *limit)
		case models.EntityCharacter:
			mails, err = r.Services.MostValuable(ctx, "victim.characterID", uint64(*id), *age, *limit)
		case models.EntityCorporation:
			mails, err = r.Services.MostValuable(ctx, "victim.corporationID", uint64(*id), *age, *limit)
		case models.EntityAlliance:
			mails, err = r.Services.MostValuable(ctx, "victim.allianceID", uint64(*id), *age, *limit)
		case models.EntityShip:
			mails, err = r.Services.MostValuable(ctx, "victim.ShipTypeID", uint64(*id), *age, *limit)
		case models.EntityShipGroup:
			mails, err = r.Services.MostValuable(ctx, "victim.ShipGroupID", uint64(*id), *age, *limit)
		default:
			return nil, errors.New("invalid entity")
		}
	default:
		return nil, errors.New("invalid category specified")
	}

	return mails, err

}

func (r *queryResolver) KillmailsByEntityID(ctx context.Context, entity models.Entity, id int, page *int, filter *models.KillmailFilter) ([]*neo.Killmail, error) {

	var mails []*neo.Killmail

	ops, err := buildOperators(filter)
	if err != nil {
		return nil, err
	}

	switch entity {
	case models.EntityAll:
		return nil, errors.New("All Type is not supported on this query")
	case models.EntityCharacter:
		mails, err = r.Services.KillmailsByCharacterID(ctx, uint64(id), uint(*page), ops...)
	case models.EntityCorporation:
		mails, err = r.Services.KillmailsByCorporationID(ctx, uint(id), uint(*page), ops...)
	case models.EntityAlliance:
		mails, err = r.Services.KillmailsByAllianceID(ctx, uint(id), uint(*page), ops...)
	case models.EntityShip:
		mails, err = r.Services.KillmailsByShipID(ctx, uint(id), uint(*page), ops...)
	case models.EntityShipGroup:
		mails, err = r.Services.KillmailsByShipGroupID(ctx, uint(id), uint(*page), ops...)
	case models.EntitySystem:
		mails, err = r.Services.KillmailsBySystemID(ctx, uint(id), uint(*page), ops...)
	case models.EntityConstellation:
		mails, err = r.Services.KillmailsByConstellationID(ctx, uint(id), uint(*page), ops...)
	case models.EntityRegion:
		mails, err = r.Services.KillmailsByRegionID(ctx, uint(id), uint(*page), ops...)
	default:
		return nil, errors.New("invalid entity")
	}

	return mails, err
}

func (r *subscriptionResolver) KillmailFeed(ctx context.Context) (<-chan *neo.Killmail, error) {
	feedCh := make(chan *neo.Killmail)

	go func(ctx context.Context, output chan *neo.Killmail) {
		r.Logger.Println("hello")

		feed := r.Redis.Subscribe(ctx, "killmail-feed")

		_, err := feed.Receive(ctx)
		if err != nil {
			return
		}

		subCh := feed.ChannelSize(1)
		r.Logger.Info("successfully created feedCh")

		for {

			select {
			case <-ctx.Done():
				close(output)
				feed.Close()
				r.Logger.Info("ctx.Done, hanging up")
				return
			case msg := <-subCh:
				var killmail = new(neo.Killmail)
				err = json.Unmarshal([]byte(msg.Payload), killmail)
				if err != nil {
					r.Logger.WithError(err).Error("failed to unmarshal feed payload")
				}
				r.Logger.WithField("id", killmail.ID).WithField("hash", killmail.Hash).Info("pushing message to feed")
				output <- killmail
				break
			}
		}

	}(ctx, feedCh)

	return feedCh, err

}

func (r *Resolver) Killmail() service.KillmailResolver {
	return &killmailResolver{r}
}

type killmailResolver struct{ *Resolver }

func (r *killmailResolver) System(ctx context.Context, obj *neo.Killmail) (*neo.SolarSystem, error) {
	return r.Dataloader(ctx).SolarSystemLoader.Load(obj.SolarSystemID)
}

func (r *killmailResolver) Attackers(ctx context.Context, obj *neo.Killmail, finalBlowOnly *bool) ([]*neo.KillmailAttacker, error) {
	if *finalBlowOnly {
		for _, attacker := range obj.Attackers {
			if attacker.FinalBlow {
				return []*neo.KillmailAttacker{attacker}, nil
			}
		}
		return nil, nil
	}

	return obj.Attackers, nil
}
