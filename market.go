package neo

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
)

type MarketRepository interface {
	BuiltPrice(ctx context.Context, id uint64, date time.Time) (*PriceBuilt, error)
	InsertBuiltPrice(ctx context.Context, price *PriceBuilt) (*PriceBuilt, error)
	HistoricalRecord(ctx context.Context, id uint64, date time.Time, limit null.Int) ([]*HistoricalRecord, error)
	CreateHistoricalRecord(ctx context.Context, records []*HistoricalRecord) ([]*HistoricalRecord, error)

	AvgOfTypeLowPrice(ctx context.Context, id uint64, days int, date time.Time) (null.Float64, error)
}

type HistoricalRecord struct {
	TypeID uint64  `db:"type_id" json:"typeID"`
	Date   *Date   `db:"date" json:"date"`
	Price  float64 `db:"price" json:"average"`
}

type PriceBuilt struct {
	TypeID uint64    `db:"type_id" json:"typeID"`
	Date   time.Time `db:"date" json:"date"`
	Price  float64   `db:"price" json:"price"`
}

type MarketPrices struct {
	AdjustedPrice float64 `json:"adjusted_price"`
	AveragePrice  float64 `json:"average_price"`
	TypeID        uint64  `json:"type_id"`
}

type MarketGroup struct {
	MarketGroupID uint64   `json:"market_group_id"`
	ParentGroupID uint64   `json:"parent_group_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Types         []uint64 `json:"types"`
}

type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(data []byte) error {

	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	*d = Date{t}

	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {

	return []byte(d.Format("2016-01-02")), nil

}

func (d *Date) Scan(v interface{}) error {

	if v == nil {
		*d = Date{time.Now()}
		return nil
	}

	switch v := v.(type) {
	case string:
		t, e := time.Parse("2006-01-02", v)
		if e != nil {
			return e
		}

		*d = Date{t}
		return nil
	}

	return nil

}

func (d *Date) Value() (driver.Value, error) {
	return d.Format("2006-01-02"), nil
}
