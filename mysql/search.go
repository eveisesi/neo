package mysql

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
)

type searchRespository struct {
	db *sqlx.DB
}

func NewSearchRepository(db *sqlx.DB) neo.SearchRepository {
	return &searchRespository{db}
}

func (r *searchRespository) AllSearchableEntities(ctx context.Context) ([]*neo.SearchableEntity, error) {

	query := `
		SELECT id, name, 'characters' as type, CONCAT("characters/", id, "/portrait") AS image FROM characters
		UNION
		SELECT id, name, 'corporations' as type,CONCAT("corporations/", id, "/logo") AS image FROM corporations
		UNION
		SELECT id, name,'alliances' as type, CONCAT("alliances/", id, "/logo") AS image FROM alliances
		UNION
		(
			SELECT 
				t.id, t.name, 'ship' as type, CONCAT( "types/",t.id, "/render") AS image 
			FROM types t
			LEFT JOIN type_groups ON t.group_id = type_groups.id
			WHERE type_groups.category_id = 6
		)
	`

	entities := make([]*neo.SearchableEntity, 0)
	err := r.db.SelectContext(ctx, &entities, query)

	return entities, err

}
