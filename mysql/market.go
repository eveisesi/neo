package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/queries"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/eveisesi/neo/mysql/boiler"

	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/null"
)

type marketRepository struct {
	db *sqlx.DB
}

func NewMarketRepository(db *sqlx.DB) neo.MarketRepository {
	return &marketRepository{db}
}

func (r *marketRepository) HistoricalRecord(ctx context.Context, id uint, date time.Time, limit null.Int) ([]*neo.HistoricalRecord, error) {

	mods := make([]qm.QueryMod, 0)
	mods = append(mods,
		qm.Select("type_id", "date", "price"),
		boiler.PriceWhere.TypeID.EQ(id),
		boiler.PriceWhere.Date.LTE(date),
		qm.OrderBy(boiler.PriceColumns.Date+" DESC"),
	)
	if limit.Valid {
		mods = append(mods, qm.Limit(limit.Int))
	}

	query, args := queries.BuildQuery(boiler.Prices(mods...).Query)
	records := make([]*neo.HistoricalRecord, 0)
	err := r.db.SelectContext(ctx, &records, query, args...)

	return records, err

}

func (r *marketRepository) BuiltPrice(ctx context.Context, id uint, date time.Time) (*neo.PriceBuilt, error) {

	query := `
		SELECT
			type_id,
			date,
			price
		FROM
			prices_built
		WHERE 
			type_id = ?
			AND date = ?
	`

	build := new(neo.PriceBuilt)
	err := r.db.GetContext(ctx, build, query, id, date.Format("2006-01-02"))

	return build, err

}

func (r *marketRepository) InsertBuiltPrice(ctx context.Context, price *neo.PriceBuilt) (*neo.PriceBuilt, error) {

	query := `
		INSERT INTO prices_built (
			type_id,
			date,
			price,
			created_at,
			updated_at
		) VALUES (
			?,?,?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE price=VALUES(price)
	`

	_, err := r.db.ExecContext(ctx, query, price.TypeID, price.Date, price.Price)

	return price, err

}

func (r *marketRepository) CreateHistoricalRecord(ctx context.Context, records []*neo.HistoricalRecord) ([]*neo.HistoricalRecord, error) {

	query := `
		INSERT INTO prices (
			type_id,
			date,
			price,
			created_at,
			updated_at
		) VALUES %s ON DUPLICATE KEY UPDATE price=VALUES(price)
	`
	args := make([]string, 0)
	params := make([]interface{}, 0)
	for _, record := range records {
		args = append(args, "(?, ?, ?, NOW(), NOW())")
		params = append(params, record.TypeID, record.Date, record.Price)
	}

	query = fmt.Sprintf(query, strings.Join(args, ","))

	_, err := r.db.ExecContext(ctx, query, params...)

	return records, err

}

func (r *marketRepository) AvgOfTypeLowPrice(ctx context.Context, id uint, days int, date time.Time) (null.Float64, error) {

	query := `
		SELECT 
			AVG(average) 
		FROM 
			prices 
		WHERE 
			type_id = ? 
			AND date BETWEEN ? - INTERVAL %d DAY AND ?

	`

	query = fmt.Sprintf(query, days)

	var avg null.Float64

	err := r.db.GetContext(ctx, &avg, query, id, date, date)

	return avg, err

}
