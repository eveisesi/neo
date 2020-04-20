package neo

import (
	"context"
	"time"
)

type BlueprintRepository interface {
	BlueprintMaterials(context.Context, uint64) ([]*BlueprintMaterial, error)
	BlueprintProduct(context.Context, uint64) (*BlueprintProduct, error)
	BlueprintProductByProductTypeID(context.Context, uint64) (*BlueprintProduct, error)
}

// BlueprintMaterial is an object representing the database table.
type BlueprintMaterial struct {
	TypeID         uint64    `db:"type_id" json:"typeID"`
	ActivityID     uint64    `db:"activity_id" json:"activityID"`
	MaterialTypeID uint64    `db:"material_type_id" json:"materialTypeID"`
	Quantity       uint64    `db:"quantity" json:"quantity"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

// BlueprintProduct is an object representing the database table.
type BlueprintProduct struct {
	TypeID        uint64    `db:"type_id" json:"typeID"`
	ActivityID    uint64    `db:"activity_id" json:"activityID"`
	ProductTypeID uint64    `db:"product_type_id" json:"productTypeID"`
	Quantity      uint64    `db:"quantity" json:"quantity"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
}
