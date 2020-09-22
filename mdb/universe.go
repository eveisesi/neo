package mdb

import (
	"context"
	"time"

	"github.com/eveisesi/neo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type universeRepository struct {
	constellations *mongo.Collection // Constellation
	regions        *mongo.Collection // Region
	systems        *mongo.Collection // Systems
	items          *mongo.Collection // Types
	attributes     *mongo.Collection // Type Attributes
	categories     *mongo.Collection // Type Categories
	flags          *mongo.Collection // Type Flags
	groups         *mongo.Collection // Type Groups
}

func NewUniverseRepository(d *mongo.Database) neo.UniverseRepository {
	return &universeRepository{
		d.Collection("constellations"),
		d.Collection("regions"),
		d.Collection("systems"),
		d.Collection("types"),
		d.Collection("typeAttributes"),
		d.Collection("typeCategories"),
		d.Collection("typeFlags"),
		d.Collection("typeGroups"),
	}
}

func (r *universeRepository) Constellation(ctx context.Context, id uint) (*neo.Constellation, error) {

	var constellation = new(neo.Constellation)

	err := r.constellations.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(constellation)

	return constellation, err

}

func (r *universeRepository) Constellations(ctx context.Context, operators ...*neo.Operator) ([]*neo.Constellation, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var constellations = make([]*neo.Constellation, 0)
	result, err := r.constellations.Find(ctx, filters, options)
	if err != nil {
		return constellations, err
	}

	err = result.All(ctx, &constellations)

	return constellations, err

}

func (r *universeRepository) Region(ctx context.Context, id uint) (*neo.Region, error) {

	var region = new(neo.Region)

	err := r.regions.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(region)

	return region, err

}

func (r *universeRepository) Regions(ctx context.Context, operators ...*neo.Operator) ([]*neo.Region, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var regions = make([]*neo.Region, 0)
	result, err := r.regions.Find(ctx, filters, options)
	if err != nil {
		return regions, err
	}

	err = result.All(ctx, &regions)

	return regions, err

}

func (r *universeRepository) SolarSystem(ctx context.Context, id uint) (*neo.SolarSystem, error) {

	var system = new(neo.SolarSystem)

	err := r.systems.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(system)

	return system, err

}

func (r *universeRepository) CreateSolarSystem(ctx context.Context, system *neo.SolarSystem) error {

	system.CreatedAt = time.Now().Unix()
	system.UpdatedAt = time.Now().Unix()

	_, err := r.systems.InsertOne(ctx, system)

	return err

}

func (r *universeRepository) SolarSystems(ctx context.Context, operators ...*neo.Operator) ([]*neo.SolarSystem, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var systems = make([]*neo.SolarSystem, 0)
	result, err := r.systems.Find(ctx, filters, options)
	if err != nil {
		return systems, err
	}

	err = result.All(ctx, &systems)

	return systems, err

}

func (r *universeRepository) Type(ctx context.Context, id uint) (*neo.Type, error) {
	var item = new(neo.Type)

	err := r.items.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(item)

	return item, err
}

func (r *universeRepository) CreateType(ctx context.Context, item *neo.Type) error {

	item.CreatedAt = time.Now().Unix()
	item.UpdatedAt = time.Now().Unix()

	_, err := r.items.InsertOne(ctx, item)

	return err

}

func (r *universeRepository) Types(ctx context.Context, operators ...*neo.Operator) ([]*neo.Type, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var items = make([]*neo.Type, 0)
	result, err := r.items.Find(ctx, filters, options)
	if err != nil {
		return items, err
	}

	err = result.All(ctx, &items)

	return items, err

}

func (r *universeRepository) CreateTypeAttributes(ctx context.Context, attributes []*neo.TypeAttribute) error {

	attrInterface := make([]interface{}, len(attributes))
	for i, attribute := range attributes {
		attribute.CreatedAt = time.Now().Unix()
		attribute.UpdatedAt = time.Now().Unix()
		attrInterface[i] = attribute
	}

	_, err := r.items.InsertMany(ctx, attrInterface)

	return err

}

func (r *universeRepository) TypeAttributes(ctx context.Context, operators ...*neo.Operator) ([]*neo.TypeAttribute, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var attributes = make([]*neo.TypeAttribute, 0)
	result, err := r.attributes.Find(ctx, filters, options)
	if err != nil {
		return attributes, err
	}

	err = result.All(ctx, &attributes)

	return attributes, err

}

func (r *universeRepository) TypeCategory(ctx context.Context, id uint) (*neo.TypeCategory, error) {
	var category = new(neo.TypeCategory)

	err := r.categories.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(category)

	return category, err
}

func (r *universeRepository) TypeCategories(ctx context.Context, operators ...*neo.Operator) ([]*neo.TypeCategory, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var categories = make([]*neo.TypeCategory, 0)
	result, err := r.categories.Find(ctx, filters, options)
	if err != nil {
		return categories, err
	}

	err = result.All(ctx, &categories)

	return categories, err

}

func (r *universeRepository) TypeFlag(ctx context.Context, id uint) (*neo.TypeFlag, error) {
	var flag = new(neo.TypeFlag)

	err := r.flags.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(flag)

	return flag, err
}

func (r *universeRepository) TypeFlags(ctx context.Context, operators ...*neo.Operator) ([]*neo.TypeFlag, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var flags = make([]*neo.TypeFlag, 0)
	result, err := r.flags.Find(ctx, filters, options)
	if err != nil {
		return flags, err
	}

	err = result.All(ctx, &flags)

	return flags, err

}

func (r *universeRepository) TypeGroup(ctx context.Context, id uint) (*neo.TypeGroup, error) {
	var group = new(neo.TypeGroup)

	err := r.groups.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(group)

	return group, err
}

func (r *universeRepository) TypeGroups(ctx context.Context, operators ...*neo.Operator) ([]*neo.TypeGroup, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var groups = make([]*neo.TypeGroup, 0)
	result, err := r.groups.Find(ctx, filters, options)
	if err != nil {
		return groups, err
	}

	err = result.All(ctx, &groups)

	return groups, err

}
