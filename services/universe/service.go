package universe

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/go-redis/redis/v7"
)

type Service interface {
	BlueprintMaterials(context.Context, uint64) ([]*neo.BlueprintMaterial, error)
	BlueprintProduct(context.Context, uint64) (*neo.BlueprintProduct, error)
	BlueprintProductByProductTypeID(context.Context, uint64) (*neo.BlueprintProduct, error)

	Constellation(ctx context.Context, id uint64) (*neo.Constellation, error)
	ConstellationsByConstellationIDs(ctx context.Context, ids []uint64) ([]*neo.Constellation, error)

	Region(ctx context.Context, id uint64) (*neo.Region, error)
	RegionsByRegionIDs(ctx context.Context, ids []uint64) ([]*neo.Region, error)

	SolarSystem(ctx context.Context, id uint64) (*neo.SolarSystem, error)
	SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*neo.SolarSystem, error)

	Type(ctx context.Context, id uint64) (*neo.Type, error)
	TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.Type, error)

	TypeAttributes(ctx context.Context, id uint64) ([]*neo.TypeAttribute, error)
	TypeAttributesByTypeIDs(ctx context.Context, ids []uint64) ([]*neo.TypeAttribute, error)

	TypeCategory(ctx context.Context, id uint64) (*neo.TypeCategory, error)
	TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint64) ([]*neo.TypeCategory, error)

	TypeFlag(ctx context.Context, id uint64) (*neo.TypeFlag, error)
	TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint64) ([]*neo.TypeFlag, error)

	TypeGroup(ctx context.Context, id uint64) (*neo.TypeGroup, error)
	TypeGroupsByGroupIDs(ctx context.Context, ids []uint64) ([]*neo.TypeGroup, error)
}

type service struct {
	redis *redis.Client
	esi   esi.Service
	neo.BlueprintRepository
	neo.UniverseRepository
}

func NewService(redis *redis.Client, esi esi.Service, blueprint neo.BlueprintRepository, universe neo.UniverseRepository) Service {
	return &service{
		redis,
		esi,
		blueprint,
		universe,
	}
}
