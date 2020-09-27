package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type MarketRepository interface {
	BuiltPrice(ctx context.Context, id uint, date time.Time) (*PriceBuilt, error)
	InsertBuiltPrice(ctx context.Context, price *PriceBuilt) (*PriceBuilt, error)
	HistoricalRecord(ctx context.Context, id uint, date time.Time, limit null.Int) ([]*HistoricalRecord, error)
	CreateHistoricalRecord(ctx context.Context, records []*HistoricalRecord) ([]*HistoricalRecord, error)

	Price(ctx context.Context, typeID uint, date string) (*HistoricalRecord, error)
	Prices(ctx context.Context, operators ...*Operator) ([]*HistoricalRecord, error)
}

type HistoricalRecord struct {
	TypeID uint    `bson:"typeID" json:"typeID"`
	Date   string  `bson:"date" json:"date"`
	Price  float64 `bson:"price" json:"average"`
}

type PriceBuilt struct {
	TypeID uint    `bson:"typeID" json:"typeID"`
	Date   string  `bson:"date" json:"date"`
	Price  float64 `bson:"price" json:"price"`
}

type MarketPrices struct {
	AdjustedPrice float64 `bson:"adjustedPrice" json:"adjustedPrice"`
	AveragePrice  float64 `bson:"averagePrice" json:"averagePrice"`
	TypeID        uint    `bson:"typeID" json:"typeID"`
}

type MarketGroup struct {
	MarketGroupID uint   `bson:"marketGroupID" json:"marketGroupID"`
	ParentGroupID uint   `bson:"parentGroupID" json:"parentGroupID"`
	Name          string `bson:"name" json:"name"`
	Description   string `bson:"description" json:"description"`
	Types         []uint `bson:"types" json:"types"`
}
