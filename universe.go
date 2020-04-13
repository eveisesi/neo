package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type UniverseRepository interface {
	SolarSystem(ctx context.Context, id uint64) (*SolarSystem, error)
	SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*SolarSystem, error)
	Type(ctx context.Context, id uint64) (*Type, error)
	TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*Type, error)
	TypeAttributes(context.Context, uint64) ([]*TypeAttribute, error)
	TypeAttributesByTypeIDs(context.Context, []uint64) ([]*TypeAttribute, error)
	TypeCategory(context.Context, uint64) (*TypeCategory, error)
	TypeCategoriesByCategoryIDs(context.Context, []uint64) ([]*TypeCategory, error)
	TypeFlag(context.Context, uint64) (*TypeFlag, error)
	TypeFlagsByTypeFlagIDs(context.Context, []uint64) ([]*TypeFlag, error)
	TypeGroup(context.Context, uint64) (*TypeGroup, error)
	TypeGroupsByGroupIDs(context.Context, []uint64) ([]*TypeGroup, error)
}

// Type is an object representing the database table.
type Type struct {
	ID            uint64       `json:"id"`
	GroupID       uint64       `json:"groupID"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Volume        float64      `json:"volume"`
	RaceID        null.Uint64  `json:"raceID"`
	BasePrice     null.Float64 `json:"base_price"`
	Published     bool         `json:"published"`
	MarketGroupID null.Uint64  `json:"marketGroupID"`
	CreatedAt     null.Time    `json:"created_at"`
	UpdatedAt     null.Time    `json:"updated_at"`
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
	ID        uint64    `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name      string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Published null.Bool `boil:"published" json:"published,omitempty" toml:"published" yaml:"published,omitempty"`
	CreatedAt time.Time `boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
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
	ID         uint64    `boil:"id" json:"id" toml:"id" yaml:"id"`
	CategoryID uint64    `boil:"category_id" json:"categoryID" toml:"categoryID" yaml:"categoryID"`
	Name       string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Published  bool      `boil:"published" json:"published" toml:"published" yaml:"published"`
	CreatedAt  time.Time `boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`
	UpdatedAt  time.Time `boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
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
