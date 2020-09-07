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

type SolarSystem struct {
	ID              uint   `bson:"id" json:"id"`
	Name            string `bson:"name" json:"name"`
	RegionID        uint   `bson:"regionID" json:"regionID"`
	ConstellationID uint   `bson:"constellationID" json:"constellationID"`
	FactionID       *int   `bson:"factionID" json:"factionID"`
	SunTypeID       *int   `bson:"sunTypeID" json:"sunTypeID"`
	Position        struct {
		X float64 `bson:"x" json:"x"`
		Y float64 `bson:"y" json:"y"`
		Z float64 `bson:"z" json:"z"`
	} `bson:"position" json:"position"`
	Security  float64 `bson:"security" json:"security"`
	CreatedAt int64   `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64   `bson:"updatedAt" json:"updatedAt"`

	Constellation *Constellation `bson:"-" json:"-"`
}

// Constellation is an object representing the database table.
type Constellation struct {
	ID       uint   `bson:"id" json:"id"`
	Name     string `bson:"name" json:"name"`
	RegionID uint   `bson:"regionID" json:"regionID"`
	Position struct {
		X float64 `bson:"x" json:"x"`
		Y float64 `bson:"y" json:"y"`
		Z float64 `bson:"z" json:"z"`
	} `bson:"position" json:"position"`
	FactionID *int  `bson:"factionID" json:"factionID"`
	CreatedAt int64 `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64 `bson:"updatedAt" json:"updatedAt"`

	Region *Region `bson:"-" json:"-"`
}

// Region is an object representing the database table.
type Region struct {
	ID       uint   `bson:"id" json:"id"`
	Name     string `bson:"name" json:"name"`
	Position struct {
		X float64 `bson:"x" json:"x"`
		Y float64 `bson:"y" json:"y"`
		Z float64 `bson:"z" json:"z"`
	} `bson:"position" json:"position"`
	FactionID *uint `bson:"factionID" json:"factionID"`
	CreatedAt int64 `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64 `bson:"updatedAt" json:"updatedAt"`
}

// Type is an object representing the database table.
type Type struct {
	ID            uint   `bson:"id" json:"id"`
	GroupID       uint   `bson:"groupID" json:"groupID"`
	Name          string `bson:"name" json:"name"`
	Description   string `bson:"description" json:"description"`
	Published     bool   `bson:"published" json:"published"`
	MarketGroupID *uint  `bson:"marketGroupID" json:"marketGroupID"`
	CreatedAt     int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt     int64  `bson:"updatedAt" json:"updatedAt"`

	Group *TypeGroup `bson:"-" json:"-"`
}

// TypeAttribute is an object representing the database table.
type TypeAttribute struct {
	TypeID      uint  `bson:"typeID" json:"typeID"`
	AttributeID uint  `bson:"attributeID" json:"attributeID"`
	Value       int64 `bson:"value" json:"value"`
	CreatedAt   int64 `bson:"createdAt" json:"createdAt"`
	UpdatedAt   int64 `bson:"updatedAt" json:"updatedAt"`
}

// TypeCategory is an object representing the database table.
type TypeCategory struct {
	ID        uint   `bson:"id" json:"id"`
	Name      string `bson:"name" json:"name"`
	Published bool   `bson:"published" json:"published"`
	CreatedAt int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt" json:"updatedAt"`
}

// TypeFlag is an object representing the database table.
type TypeFlag struct {
	ID        uint   `bson:"id" json:"id"`
	Name      string `bson:"name" json:"name"`
	Text      string `bson:"text" json:"text"`
	CreatedAt int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt" json:"updatedAt"`
}

// TypeGroup is an object representing the database table.
type TypeGroup struct {
	ID         uint   `bson:"id" json:"id"`
	CategoryID uint   `bson:"categoryID" json:"categoryID"`
	Name       string `bson:"name" json:"name"`
	Published  bool   `bson:"published" json:"published"`
	CreatedAt  int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt  int64  `bson:"updatedAt" json:"updatedAt"`
}

// TypeMaterial is an object representing the database table.
type TypeMaterial struct {
	TypeID         uint `bson:"typeID" json:"typeID"`
	MaterialTypeID uint `bson:"materialTypeID" json:"materialTypeID"`
	Quantity       uint `bson:"quantity" json:"quantity"`
}
