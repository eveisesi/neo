package mysql

import (
	"context"
	"fmt"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type universeRepository struct {
	db *sqlx.DB
}

func NewUniverseRepository(db *sqlx.DB) neo.UniverseRepository {
	return &universeRepository{
		db,
	}
}

func (r *universeRepository) SolarSystem(ctx context.Context, id uint64) (*neo.SolarSystem, error) {

	var solarSystem = neo.SolarSystem{}
	err := boiler.SolarSystems(
		boiler.SolarSystemWhere.ID.EQ(id),
	).Bind(ctx, r.db, &solarSystem)

	return &solarSystem, err

}

func (r *universeRepository) SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*neo.SolarSystem, error) {

	var systems = make([]*neo.SolarSystem, 0)
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

func (r *universeRepository) Type(ctx context.Context, id uint64) (*neo.Type, error) {

	var invType = neo.Type{}
	err := boiler.Types(
		boiler.TypeWhere.ID.EQ(id),
	).Bind(ctx, r.db, &invType)

	return &invType, err

}

func (r *universeRepository) TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.Type, error) {

	var invTypes = make([]*neo.Type, 0)
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

func (r *universeRepository) TypeAttributes(ctx context.Context, id uint64) ([]*neo.TypeAttribute, error) {

	attributes := make([]*neo.TypeAttribute, 0)
	err := boiler.TypeAttributes(
		boiler.TypeAttributeWhere.TypeID.EQ(id),
	).Bind(ctx, r.db, &attributes)

	return attributes, err
}

func (r *universeRepository) TypeAttributesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.TypeAttribute, error) {

	attributes := make([]*neo.TypeAttribute, 0)
	err := boiler.TypeAttributes(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.TypeAttributeColumns.TypeID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &attributes)

	return attributes, err
}

func (r *universeRepository) TypeCategory(ctx context.Context, id uint64) (*neo.TypeCategory, error) {

	category := neo.TypeCategory{}
	err := boiler.TypeCategories(
		boiler.TypeCategoryWhere.ID.EQ(id),
	).Bind(ctx, r.db, &category)

	return &category, err

}

func (r *universeRepository) TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint64) ([]*neo.TypeCategory, error) {

	categories := make([]*neo.TypeCategory, 0)
	err := boiler.TypeCategories(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.TypeCategoryColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &categories)

	return categories, err

}

func (r *universeRepository) TypeFlag(ctx context.Context, id uint64) (*neo.TypeFlag, error) {

	flag := neo.TypeFlag{}
	err := boiler.TypeFlags(
		boiler.TypeFlagWhere.ID.EQ(id),
	).Bind(ctx, r.db, &flag)

	return &flag, err

}

func (r *universeRepository) TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint64) ([]*neo.TypeFlag, error) {

	flags := make([]*neo.TypeFlag, 0)
	err := boiler.TypeFlags(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.TypeFlagColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &flags)

	return flags, err

}

func (r *universeRepository) TypeGroup(ctx context.Context, id uint64) (*neo.TypeGroup, error) {

	group := neo.TypeGroup{}
	err := boiler.TypeGroups(
		boiler.TypeGroupWhere.ID.EQ(id),
	).Bind(ctx, r.db, &group)

	return &group, err

}

func (r *universeRepository) TypeGroupsByGroupIDs(ctx context.Context, ids []uint64) ([]*neo.TypeGroup, error) {

	groups := make([]*neo.TypeGroup, 0)
	err := boiler.TypeGroups(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.TypeGroupColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &groups)

	return groups, err
}
