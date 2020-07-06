package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
)

type statRepository struct {
	db *sqlx.DB
}

func NewStatRepository(db *sqlx.DB) neo.StatsRepository {
	return &statRepository{
		db,
	}
}

func (r *statRepository) Apply(id uint64, entity neo.StatEntity, category neo.StatCategory, frequency neo.StatFrequency, date time.Time, value float64) error {

	query := `
		INSERT INTO stats (id, entity, category, frequency, date, value, created_at, updated_at) 
		VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			NOW(),
			NOW()
		) ON DUPLICATE KEY UPDATE value = VALUES(value) + value

	`

	_, err := r.db.Exec(query, id, entity.String(), category.String(), frequency.String(), date, value)
	return err

}

func (r *statRepository) Save(ctx context.Context, stats []*neo.Stat) error {

	places := make([]string, 0)
	values := make([]interface{}, 0)

	for _, v := range stats {
		places = append(places, "(?, ?, ?, ?, ?, ?, NOW(), NOW())")
		values = append(values, v.ID, v.Entity, v.Category, v.Frequency, v.Date, v.Value)
	}

	query := `
		INSERT INTO stats (id, entity, category, frequency, date, value, created_at, updated_at) 
		VALUES %s ON DUPLICATE KEY UPDATE value = VALUES(value) + value
	`

	query = fmt.Sprintf(query, strings.Join(places, ","))

	_, err := r.db.ExecContext(ctx, query, values...)

	return err

}
