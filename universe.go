package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type UniverseRepository interface {
	Constellation(ctx context.Context, id uint) (*Constellation, error)
	ConstellationsByConstellationIDs(ctx context.Context, ids []uint) ([]*Constellation, error)

	Region(ctx context.Context, id uint) (*Region, error)
	RegionsByRegionIDs(ctx context.Context, ids []uint) ([]*Region, error)

	SolarSystem(ctx context.Context, id uint) (*SolarSystem, error)
	CreateSolarSystem(ctx context.Context, system *SolarSystem) error
	SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint) ([]*SolarSystem, error)

	Type(ctx context.Context, id uint) (*Type, error)
	CreateType(ctx context.Context, invType *Type) error
	TypesByTypeIDs(ctx context.Context, ids []uint) ([]*Type, error)

	TypeAttributes(ctx context.Context, id uint) ([]*TypeAttribute, error)
	CreateTypeAttributes(ctx context.Context, attributes []*TypeAttribute) error
	TypeAttributesByTypeIDs(ctx context.Context, ids []uint) ([]*TypeAttribute, error)

	TypeCategory(ctx context.Context, id uint) (*TypeCategory, error)
	TypeCategoriesByCategoryIDs(ctx context.Context, ids []uint) ([]*TypeCategory, error)

	TypeFlag(ctx context.Context, id uint) (*TypeFlag, error)
	TypeFlagsByTypeFlagIDs(ctx context.Context, ids []uint) ([]*TypeFlag, error)

	TypeGroup(ctx context.Context, id uint) (*TypeGroup, error)
	TypeGroupsByGroupIDs(ctx context.Context, ids []uint) ([]*TypeGroup, error)
}

// Constellation is an object representing the database table.
type Constellation struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	RegionID  uint       `json:"regionID"`
	PosX      float64    `json:"posX"`
	PosY      float64    `json:"posY"`
	PosZ      float64    `json:"posZ"`
	FactionID null.Int64 `json:"factionID"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`

	Region *Region `json:"-"`
}

// Region is an object representing the database table.
type Region struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	PosX      float64   `json:"posX"`
	PosY      float64   `json:"posY"`
	PosZ      float64   `json:"posZ"`
	FactionID null.Uint `json:"factionID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SolarSystem struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	RegionID        uint       `json:"regionID"`
	ConstellationID uint       `json:"constellationID"`
	FactionID       null.Int64 `json:"factionID"`
	SunTypeID       null.Int64 `json:"sun_typeID"`
	PosX            float64    `json:"pos_x"`
	PosY            float64    `json:"pos_y"`
	PosZ            float64    `json:"pos_z"`
	Security        float64    `json:"security"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Constellation *Constellation `json:"-"`
}

// Type is an object representing the database table.
type Type struct {
	ID            uint      `json:"id"`
	GroupID       uint      `json:"groupID"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Published     bool      `json:"published"`
	MarketGroupID null.Uint `json:"marketGroupID"`
	CreatedAt     null.Time `json:"created_at"`
	UpdatedAt     null.Time `json:"updated_at"`

	Group *TypeGroup `json:"-"`
}

// TypeAttribute is an object representing the database table.
type TypeAttribute struct {
	TypeID      uint      `json:"typeID"`
	AttributeID uint      `json:"attributeID"`
	Value       int64     `json:"value"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TypeCategory is an object representing the database table.
type TypeCategory struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Published null.Bool `json:"published"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TypeFlag is an object representing the database table.
type TypeFlag struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TypeGroup is an object representing the database table.
type TypeGroup struct {
	ID         uint      `json:"id"`
	CategoryID uint      `json:"categoryID"`
	Name       string    `json:"name"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TypeMaterial is an object representing the database table.
type TypeMaterial struct {
	TypeID         uint `boil:"type_id" json:"typeID" toml:"typeID" yaml:"typeID"`
	MaterialTypeID uint `boil:"material_type_id" json:"materialTypeID" toml:"materialTypeID" yaml:"materialTypeID"`
	Quantity       uint `boil:"quantity" json:"quantity" toml:"quantity" yaml:"quantity"`
}
