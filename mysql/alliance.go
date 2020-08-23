package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

type allianceRepository struct {
	db *sqlx.DB
}

func NewAllianceRepository(db *sqlx.DB) neo.AllianceRespository {
	return &allianceRepository{
		db,
	}
}

func (r *allianceRepository) Alliance(ctx context.Context, id uint) (*neo.Alliance, error) {

	var alliance = neo.Alliance{}
	err := boiler.Alliances(
		boiler.AllianceWhere.ID.EQ(id),
	).Bind(ctx, r.db, &alliance)

	return &alliance, err

}

func (r *allianceRepository) Alliances(ctx context.Context, mods ...neo.Modifier) ([]*neo.Alliance, error) {

	if len(mods) == 0 {
		return nil, fmt.Errorf("atleast one modifier must be passed in")
	}

	alliances := make([]*neo.Alliance, 0)
	err := boiler.Alliances(BuildQueryModifiers(boiler.TableNames.Alliances, mods...)...).Bind(ctx, r.db, &alliances)
	return alliances, err

}

func (r *allianceRepository) Expired(ctx context.Context) ([]*neo.Alliance, error) {

	mods := []neo.Modifier{
		neo.LessThanTime{Column: "CacheUntil", Value: time.Now()},
		neo.LimitModifier(1000),
		neo.OrderModifier{Column: "CacheUntil", Sort: neo.SortAsc},
	}

	return r.Alliances(ctx, mods...)

}

func (r *allianceRepository) CreateAlliance(ctx context.Context, alliance *neo.Alliance) (*neo.Alliance, error) {

	var bAlliance = new(boiler.Alliance)
	err := copier.Copy(bAlliance, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to copy alliance to orm")
	}

	err = bAlliance.Insert(ctx, r.db, boil.Infer(), true)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance into db")
	}

	err = copier.Copy(alliance, bAlliance)

	return alliance, errors.Wrap(err, "unable to copy orm to alliance")

}

func (r *allianceRepository) UpdateAlliance(ctx context.Context, id uint, alliance *neo.Alliance) (*neo.Alliance, error) {

	var bAlliance = new(boiler.Alliance)
	err := copier.Copy(bAlliance, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to copy alliance to orm")
	}

	bAlliance.ID = id

	_, err = bAlliance.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance in db")
	}

	err = copier.Copy(alliance, bAlliance)

	return alliance, errors.Wrap(err, "unable to copy orm to alliance")

}

func (r *allianceRepository) MemberCountByAllianceID(ctx context.Context, id uint) (int, error) {

	query := `
		SELECT SUM(member_count) as count FROM corporations where alliance_id = ?
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, id)

	return count, err

}
