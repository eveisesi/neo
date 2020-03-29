package mysql

import (
	"context"
	"fmt"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type corporationRepository struct {
	db *sqlx.DB
}

func NewCorporationRepository(db *sqlx.DB) neo.CorporationRespository {
	return &corporationRepository{
		db,
	}
}

func (r *corporationRepository) Corporation(ctx context.Context, id uint64) (*neo.Corporation, error) {

	var corporation = neo.Corporation{}
	err := boiler.Corporations(
		boiler.CorporationWhere.ID.EQ(id),
	).Bind(ctx, r.db, &corporation)

	return &corporation, err

}

func (r *corporationRepository) CorporationsByCorporationIDs(ctx context.Context, ids []uint64) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	err := boiler.Corporations(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.CorporationColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &corporations)

	return corporations, err
}
