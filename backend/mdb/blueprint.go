package mdb

import (
	"context"

	"github.com/eveisesi/neo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type blueprintRepository struct {
	materials *mongo.Collection
	products  *mongo.Collection
}

func NewBlueprintRepository(d *mongo.Database) neo.BlueprintRepository {
	return &blueprintRepository{
		d.Collection("blueprintMaterials"),
		d.Collection("blueprintProducts"),
	}
}

func (r *blueprintRepository) BlueprintMaterials(ctx context.Context, id uint) ([]*neo.BlueprintMaterial, error) {

	var materials = make([]*neo.BlueprintMaterial, 0)

	result, err := r.materials.Find(ctx, primitive.D{primitive.E{Key: "typeID", Value: id}, primitive.E{Key: "activityID", Value: 1}})
	if err != nil {
		return materials, err
	}

	err = result.All(ctx, &materials)
	return materials, err

}

func (r *blueprintRepository) BlueprintProduct(ctx context.Context, id uint) (*neo.BlueprintProduct, error) {

	var product = new(neo.BlueprintProduct)

	err := r.products.FindOne(ctx, primitive.D{primitive.E{Key: "typeID", Value: id}}).Decode(product)

	return product, err

}

func (r *blueprintRepository) BlueprintProductByProductTypeID(ctx context.Context, id uint) (*neo.BlueprintProduct, error) {

	var product = new(neo.BlueprintProduct)

	err := r.products.FindOne(ctx, primitive.D{primitive.E{Key: "productTypeID", Value: id}}).Decode(product)

	return product, err

}
