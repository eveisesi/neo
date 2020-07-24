package neo

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"
	"time"
)

type StatsRepository interface {
	Apply(id uint64, entity StatEntity, category StatCategory, frequency StatFrequency, date time.Time, value float64) error
	Save(ctx context.Context, stats []*Stat) error
}

// Stat is an object representing the database table.
type Stat struct {
	ID        uint64        `db:"id" json:"id"`
	Entity    StatEntity    `db:"entity" json:"entity"`
	Category  StatCategory  `db:"category" json:"category"`
	Frequency StatFrequency `db:"frequency" json:"frequency"`
	Date      *Date         `db:"date" json:"date"`
	Value     float64       `db:"value" json:"value"`
	CreatedAt time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `db:"updated_at" json:"updatedAt"`
}

type StatCategory string

const (
	StatCategoryISKKilled   StatCategory = "isk_killed"
	StatCategoryISKLost     StatCategory = "isk_lost"
	StatCategoryShipsKilled StatCategory = "ships_killed"
	StatCategoryShipsLost   StatCategory = "ships_lost"
)

var AllStatCategorys = []StatCategory{
	StatCategoryISKKilled,
	StatCategoryISKLost,
	StatCategoryShipsKilled,
	StatCategoryShipsLost,
}

func (e StatCategory) IsValid() bool {
	switch e {
	case StatCategoryISKKilled, StatCategoryISKLost, StatCategoryShipsKilled, StatCategoryShipsLost:
		return true
	}
	return false
}

func (e StatCategory) String() string {
	return string(e)
}

func (e *StatCategory) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = StatCategory(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid StatCategory", str)
	}
	return nil
}

func (e StatCategory) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *StatCategory) Scan(v interface{}) {

	switch i := v.(type) {
	case string:
		x := StatCategory(i)
		if !x.IsValid() {
			return
		}
		*e = x
	}

}

func (e *StatCategory) Value() (driver.Value, error) {
	return e.String(), nil
}

type StatEntity string

const (
	StatEntityCharacter     StatEntity = "character"
	StatEntityCorporation   StatEntity = "corporation"
	StatEntityAlliance      StatEntity = "alliance"
	StatEntityShip          StatEntity = "ship"
	StatEntitySystem        StatEntity = "system"
	StatEntityConstellation StatEntity = "constellation"
	StatEntityRegion        StatEntity = "region"
)

var AllCategories = []StatEntity{
	StatEntityCharacter,
	StatEntityCorporation,
	StatEntityAlliance,
	StatEntityShip,
	StatEntitySystem,
	StatEntityConstellation,
	StatEntityRegion,
}

func (e StatEntity) IsValid() bool {
	switch e {
	case StatEntityCharacter, StatEntityCorporation, StatEntityAlliance,
		StatEntityShip, StatEntitySystem, StatEntityConstellation, StatEntityRegion:
		return true
	}
	return false
}

func (e StatEntity) String() string {
	return string(e)
}

func (e *StatEntity) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = StatEntity(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid StatEntity", str)
	}
	return nil
}

func (e StatEntity) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *StatEntity) Scan(v interface{}) {

	switch i := v.(type) {
	case string:
		x := StatEntity(i)
		if !x.IsValid() {
			return
		}
		*e = x
	}

}

func (e *StatEntity) Value() (driver.Value, error) {
	return e.String(), nil
}

type StatFrequency string

const (
	StatFrequencyDaily   = "daily"
	StatFrequencyMonthly = "monthly"
	StatFrequencyYearly  = "yearly"
	StatFrequencyAlltime = "alltime"
)

var AllStatFrequencys = []StatFrequency{
	StatFrequencyDaily,
	StatFrequencyMonthly,
	StatFrequencyYearly,
	StatFrequencyAlltime,
}

func (e StatFrequency) IsValid() bool {
	switch e {
	case StatFrequencyDaily, StatFrequencyMonthly, StatFrequencyYearly, StatFrequencyAlltime:
		return true
	}
	return false
}

func (e StatFrequency) String() string {
	return string(e)
}

func (e *StatFrequency) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = StatFrequency(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid StatFrequency", str)
	}
	return nil
}

func (e StatFrequency) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *StatFrequency) Scan(v interface{}) {

	switch i := v.(type) {
	case string:
		x := StatFrequency(i)
		if !x.IsValid() {
			return
		}
		*e = x
	}

}

func (e *StatFrequency) Value() (driver.Value, error) {
	return e.String(), nil
}
