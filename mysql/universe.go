package mysql

import (
	"context"
	"fmt"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
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

func (r *universeRepository) Constellation(ctx context.Context, id uint64) (*neo.Constellation, error) {

	constellation := new(neo.Constellation)
	err := boiler.Constellations(
		boiler.ConstellationWhere.ID.EQ(id),
	).Bind(ctx, r.db, constellation)

	return constellation, err

}

func (r *universeRepository) ConstellationsByConstellationIDs(ctx context.Context, ids []uint64) ([]*neo.Constellation, error) {

	constellations := make([]*neo.Constellation, 0)
	err := boiler.Constellations(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.ConstellationColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &constellations)

	return constellations, err

}

func (r *universeRepository) Region(ctx context.Context, id uint64) (*neo.Region, error) {

	region := new(neo.Region)
	err := boiler.Regions(
		boiler.RegionWhere.ID.EQ(id),
	).Bind(ctx, r.db, region)

	return region, err

}

func (r *universeRepository) RegionsByRegionIDs(ctx context.Context, ids []uint64) ([]*neo.Region, error) {

	regions := make([]*neo.Region, 0)
	err := boiler.Regions(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.RegionColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &regions)

	return regions, err

}

func (r *universeRepository) SolarSystem(ctx context.Context, id uint64) (*neo.SolarSystem, error) {

	system := neo.SolarSystem{}
	err := boiler.SolarSystems(
		boiler.SolarSystemWhere.ID.EQ(id),
	).Bind(ctx, r.db, &system)

	return &system, err

}

func (r *universeRepository) CreateSolarSystem(ctx context.Context, system *neo.SolarSystem) error {

	bSolar := new(boiler.SolarSystem)
	err := copier.Copy(bSolar, system)
	if err != nil {
		return errors.Wrap(err, "unable to copy solar system to orm")
	}

	err = bSolar.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to insert solar system into db")
	}

	err = copier.Copy(system, bSolar)

	return errors.Wrap(err, "unable to copy orm to solar system")

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

func (r *universeRepository) CreateType(ctx context.Context, invType *neo.Type) error {

	bType := new(boiler.Type)
	err := copier.Copy(bType, invType)
	if err != nil {
		return errors.Wrap(err, "unable to copy invType to orm")
	}

	err = bType.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to insert invType into db")
	}

	err = copier.Copy(invType, bType)

	return errors.Wrap(err, "unable to copy orm to invType")

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

func (r *universeRepository) CreateTypeAttributes(ctx context.Context, attributes []*neo.TypeAttribute) error {

	txn, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "unable to start txn to insert attributes")
	}

	for _, attribute := range attributes {

		var bAttribute = new(boiler.TypeAttribute)
		err := copier.Copy(bAttribute, attribute)
		if err != nil {
			return errors.Wrap(err, "unable to copy attribute to orm")
		}

		err = bAttribute.Insert(ctx, txn, boil.Infer())
		if err != nil {
			txnErr := txn.Rollback()
			if txnErr != nil {
				err = errors.Wrap(err, "failed to rollback txn")
			}
			return errors.Wrap(err, "failed to insert type attribute into db")
		}

		err = copier.Copy(attribute, bAttribute)
		if err != nil {
			txnErr := txn.Rollback()
			if txnErr != nil {
				err = errors.Wrap(err, "failed to rollback txn")
			}
			return errors.Wrap(err, "failed to copy orm back to attribute")
		}

	}

	return errors.Wrap(txn.Commit(), "txn failed to commit")

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
