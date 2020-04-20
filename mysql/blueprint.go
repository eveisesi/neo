package mysql

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
)

type blueprintRepository struct {
	db *sqlx.DB
}

func NewBlueprintRepository(db *sqlx.DB) neo.BlueprintRepository {
	return &blueprintRepository{db}
}

func (r *blueprintRepository) BlueprintMaterials(ctx context.Context, id uint64) ([]*neo.BlueprintMaterial, error) {

	materials := make([]*neo.BlueprintMaterial, 0)
	query := `
		SELECT 
			type_id,
			activity_id,
			material_type_id,
			quantity,
			created_at,
			updated_at
		FROM 
			blueprint_materials
		WHERE
			type_id = ?
			AND activity_id = 1
	`

	err := r.db.SelectContext(ctx, &materials, query, id)

	return materials, err
}

func (r *blueprintRepository) BlueprintProduct(ctx context.Context, id uint64) (*neo.BlueprintProduct, error) {

	var product *neo.BlueprintProduct
	query := `
		SELECT 
			type_id,
			activity_id,
			product_type_id,
			quantity,
			created_at,
			updated_at
		FROM
			blueprint_products
		WHERE
			type_id = ?
	`

	err := r.db.GetContext(ctx, product, query, id)

	return product, err

}

func (r *blueprintRepository) BlueprintProductByProductTypeID(ctx context.Context, id uint64) (*neo.BlueprintProduct, error) {

	product := new(neo.BlueprintProduct)
	query := `
		SELECT 
			type_id,
			activity_id,
			product_type_id,
			quantity,
			created_at,
			updated_at
		FROM
			blueprint_products
		WHERE
			product_type_id = ?
			AND activity_id = 1
	`

	err := r.db.GetContext(ctx, product, query, id)

	return product, err

}
