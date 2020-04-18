package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type UniverseRepository interface {
	Constellation(ctx context.Context, id uint64) (*Constellation, error)
	ConstellationsByConstellationIDs(ctx context.Context, ids []uint64) ([]*Constellation, error)

	Region(ctx context.Context, id uint64) (*Region, error)
	RegionsByRegionIDs(ctx context.Context, ids []uint64) ([]*Region, error)

	SolarSystem(ctx context.Context, id uint64) (*SolarSystem, error)
	CreateSolarSystem(ctx context.Context, system *SolarSystem) error
	SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*SolarSystem, error)

	Type(ctx context.Context, id uint64) (*Type, error)
	CreateType(ctx context.Context, invType *Type) error
	TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*Type, error)

	TypeAttributes(ctx context.Context, id uint64) ([]*TypeAttribute, error)
	CreateTypeAttributes(ctx context.Context, attributes []*TypeAttribute) error
	TypeAttributesByTypeIDs(ctx context.Context, ids []uint64) ([]*TypeAttribute, error)

	TypeCategory(ctx context.Context, id uint64) (*TypeCategory, error)
	TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint64) ([]*TypeCategory, error)

	TypeFlag(ctx context.Context, id uint64) (*TypeFlag, error)
	TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint64) ([]*TypeFlag, error)

	TypeGroup(ctx context.Context, id uint64) (*TypeGroup, error)
	TypeGroupsByGroupIDs(ctx context.Context, ids []uint64) ([]*TypeGroup, error)
}

// Constellation is an object representing the database table.
type Constellation struct {
	ID        uint64     `json:"id"`
	Name      string     `json:"name"`
	RegionID  uint64     `json:"regionID"`
	PosX      float64    `json:"posX"`
	PosY      float64    `json:"posY"`
	PosZ      float64    `json:"posZ"`
	FactionID null.Int64 `json:"factionID"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Region is an object representing the database table.
type Region struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	PosX      float64   `json:"posX"`
	PosY      float64   `json:"posY"`
	PosZ      float64   `json:"posZ"`
	FactionID null.Uint `json:"factionID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SolarSystem struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	RegionID        uint64     `json:"regionID"`
	ConstellationID uint64     `json:"constellationID"`
	FactionID       null.Int64 `json:"factionID"`
	SunTypeID       null.Int64 `json:"sun_typeID"`
	PosX            float64    `json:"pos_x"`
	PosY            float64    `json:"pos_y"`
	PosZ            float64    `json:"pos_z"`
	Security        float64    `json:"security"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Type is an object representing the database table.
type Type struct {
	ID            uint64      `json:"id"`
	GroupID       uint64      `json:"groupID"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Published     bool        `json:"published"`
	MarketGroupID null.Uint64 `json:"marketGroupID"`
	CreatedAt     null.Time   `json:"created_at"`
	UpdatedAt     null.Time   `json:"updated_at"`
}

// TypeAttribute is an object representing the database table.
type TypeAttribute struct {
	TypeID      uint64    `json:"typeID"`
	AttributeID uint64    `json:"attributeID"`
	Value       int64     `json:"value"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TypeCategory is an object representing the database table.
type TypeCategory struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Published null.Bool `json:"published"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TypeFlag is an object representing the database table.
type TypeFlag struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TypeGroup is an object representing the database table.
type TypeGroup struct {
	ID         uint64    `json:"id"`
	CategoryID uint64    `json:"categoryID"`
	Name       string    `json:"name"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
