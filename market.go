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
	Prices(ctx context.Context, mods ...Modifier) ([]*HistoricalRecord, error)
}

type HistoricalRecord struct {
	TypeID uint    `bson:"typeID" json:"typeID"`
	Date   string  `bson:"date" json:"date"`
	Price  float64 `bson:"price" json:"average"`
}

type PriceBuilt struct {
	TypeID uint      `db:"type_id" json:"typeID"`
	Date   time.Time `db:"date" json:"date"`
	Price  float64   `db:"price" json:"price"`
}

type MarketPrices struct {
	AdjustedPrice float64 `json:"adjusted_price"`
	AveragePrice  float64 `json:"average_price"`
	TypeID        uint    `json:"type_id"`
}

type MarketGroup struct {
	MarketGroupID uint   `json:"market_group_id"`
	ParentGroupID uint   `json:"parent_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Types         []uint `json:"types"`
}
