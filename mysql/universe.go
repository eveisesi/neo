package mysql

import (
	"context"
	"fmt"

	"github.com/ddouglas/killboard"
	"github.com/ddouglas/killboard/mysql/boiler"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type universeRepository struct {
	db *sqlx.DB
}

func NewUniverseRepository(db *sqlx.DB) killboard.UniverseRepository {
	return &universeRepository{
		db,
	}
}

func (r *universeRepository) Type(ctx context.Context, id uint64) (*killboard.Type, error) {

	var invType = killboard.Type{}
	err := boiler.Types(
		boiler.TypeWhere.ID.EQ(id),
	).Bind(ctx, r.db, &invType)

	return &invType, err

}

func (r *universeRepository) TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*killboard.Type, error) {

	var invTypes = make([]*killboard.Type, 0)
	err := boiler.Types(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.TypeColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &invTypes)

	return invTypes, err
}

func (r *universeRepository) SolarSystem(ctx context.Context, id uint64) (*killboard.SolarSystem, error) {

	var solarSystem = killboard.SolarSystem{}
	err := boiler.SolarSystems(
		boiler.SolarSystemWhere.ID.EQ(id),
	).Bind(ctx, r.db, &solarSystem)

	return &solarSystem, err

}

func (r *universeRepository) SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*killboard.SolarSystem, error) {

	var systems = make([]*killboard.SolarSystem, 0)
	err := boiler.SolarSystems(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.SolarSystemColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &systems)

	return systems, err
}
