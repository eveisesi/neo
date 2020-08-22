package neo

import (
	"context"
	"time"
)

type BlueprintRepository interface {
	BlueprintMaterials(context.Context, uint) ([]*BlueprintMaterial, error)
	BlueprintProduct(context.Context, uint) (*BlueprintProduct, error)
	BlueprintProductByProductTypeID(context.Context, uint) (*BlueprintProduct, error)
}

// BlueprintMaterial is an object representing the database table.
type BlueprintMaterial struct {
	TypeID         uint      `db:"type_id" json:"typeID"`
	ActivityID     uint      `db:"activity_id" json:"activityID"`
	MaterialTypeID uint      `db:"material_type_id" json:"materialTypeID"`
	Quantity       uint      `db:"quantity" json:"quantity"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

// BlueprintProduct is an object representing the database table.
type BlueprintProduct struct {
	TypeID        uint      `db:"type_id" json:"typeID"`
	ActivityID    uint      `db:"activity_id" json:"activityID"`
	ProductTypeID uint      `db:"product_type_id" json:"productTypeID"`
	Quantity      uint      `db:"quantity" json:"quantity"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
}
