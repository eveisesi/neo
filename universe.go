package neo

import (
	"context"
)

type UniverseRepository interface {
	Constellation(ctx context.Context, id uint) (*Constellation, error)
	Constellations(ctx context.Context, mods ...Modifier) ([]*Constellation, error)

	Region(ctx context.Context, id uint) (*Region, error)
	Regions(ctx context.Context, mods ...Modifier) ([]*Region, error)

	SolarSystem(ctx context.Context, id uint) (*SolarSystem, error)
	SolarSystems(ctx context.Context, mods ...Modifier) ([]*SolarSystem, error)
	CreateSolarSystem(ctx context.Context, system *SolarSystem) error

	Type(ctx context.Context, id uint) (*Type, error)
	Types(ctx context.Context, mods ...Modifier) ([]*Type, error)
	CreateType(ctx context.Context, invType *Type) error

	TypeAttributes(ctx context.Context, mods ...Modifier) ([]*TypeAttribute, error)
	CreateTypeAttributes(ctx context.Context, attributes []*TypeAttribute) error

	TypeCategory(ctx context.Context, id uint) (*TypeCategory, error)
	TypeCategories(ctx context.Context, mods ...Modifier) ([]*TypeCategory, error)

	TypeFlag(ctx context.Context, id uint) (*TypeFlag, error)
	TypeFlags(ctx context.Context, mods ...Modifier) ([]*TypeFlag, error)

	TypeGroup(ctx context.Context, id uint) (*TypeGroup, error)
	TypeGroups(ctx context.Context, mods ...Modifier) ([]*TypeGroup, error)
}

// Constellation is an object representing the database table.
type Constellation struct {
	ID       uint    `db:"id" bson:"id"`
	Name     string  `db:"name" bson:"name"`
	RegionID uint    `db:"region_id" bson:"regionID"`
	PosX     float64 `db:"pos_x" bson:"-"`
	PosY     float64 `db:"pos_y" bson:"-"`
	PosZ     float64 `db:"pos_z" bson:"-"`
	Position struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
		Z float64 `bson:"z"`
	} `bson:"position"`
	FactionID *int  `db:"faction_id" bson:"factionID"`
	CreatedAt int64 `bson:"createdAt"`
	UpdatedAt int64 `bson:"updatedAt"`

	Region *Region `bson:"-"`
}

// Region is an object representing the database table.
type Region struct {
	ID       uint    `db:"id" bson:"id"`
	Name     string  `db:"name" bson:"name"`
	PosX     float64 `db:"pos_x" bson:"-"`
	PosY     float64 `db:"pos_y" bson:"-"`
	PosZ     float64 `db:"pos_z" bson:"-"`
	Position struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
		Z float64 `bson:"z"`
	} `bson:"position"`
	FactionID *uint `db:"faction_id" bson:"factionID"`
	CreatedAt int64 `bson:"createdAt"`
	UpdatedAt int64 `bson:"updatedAt"`
}

type SolarSystem struct {
	ID              uint   `bson:"id"`
	Name            string `bson:"name"`
	RegionID        uint   `db:"region_id" bson:"regionID"`
	ConstellationID uint   `db:"constellation_id" bson:"constellationID"`
	FactionID       *int   `db:"faction_id" bson:"factionID"`
	SunTypeID       *int   `db:"sun_type_id" bson:"sunTypeID"`
	Position        struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
		Z float64 `bson:"z"`
	} `bson:"position"`
	PosX      float64 `db:"pos_x" bson:"-"`
	PosY      float64 `db:"pos_y" bson:"-"`
	PosZ      float64 `db:"pos_z" bson:"-"`
	Security  float64 `db:"security" bson:"security"`
	CreatedAt int64   `bson:"createdAt"`
	UpdatedAt int64   `bson:"updatedAt"`

	Constellation *Constellation `bson:"-"`
}

// Type is an object representing the database table.
type Type struct {
	ID            uint   `bson:"id"`
	GroupID       uint   `bson:"groupID"`
	Name          string `bson:"name"`
	Description   string `bson:"description"`
	Published     bool   `bson:"published"`
	MarketGroupID *uint  `bson:"marketGroupID"`
	CreatedAt     int64  `bson:"createdAt"`
	UpdatedAt     int64  `bson:"updatedAt"`

	Group *TypeGroup `bson:"-"`
}

// TypeAttribute is an object representing the database table.
type TypeAttribute struct {
	TypeID      uint  `bson:"typeID"`
	AttributeID uint  `bson:"attributeID"`
	Value       int64 `bson:"value"`
	CreatedAt   int64 `bson:"createdAt"`
	UpdatedAt   int64 `bson:"updatedAt"`
}

// TypeCategory is an object representing the database table.
type TypeCategory struct {
	ID        uint   `bson:"id"`
	Name      string `bson:"name"`
	Published bool   `bson:"published"`
	CreatedAt int64  `bson:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt"`
}

// TypeFlag is an object representing the database table.
type TypeFlag struct {
	ID        uint   `bson:"id"`
	Name      string `bson:"name"`
	Text      string `bson:"text"`
	CreatedAt int64  `bson:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt"`
}

// TypeGroup is an object representing the database table.
type TypeGroup struct {
	ID         uint   `bson:"id"`
	CategoryID uint   `bson:"categoryID"`
	Name       string `bson:"name"`
	Published  bool   `bson:"published"`
	CreatedAt  int64  `bson:"createdAt"`
	UpdatedAt  int64  `bson:"updatedAt"`
}

// TypeMaterial is an object representing the database table.
type TypeMaterial struct {
	TypeID         uint `boil:"type_id" bson:"typeID" toml:"typeID" yaml:"typeID"`
	MaterialTypeID uint `boil:"material_type_id" bson:"materialTypeID" toml:"materialTypeID" yaml:"materialTypeID"`
	Quantity       uint `boil:"quantity" bson:"quantity" toml:"quantity" yaml:"quantity"`
}
