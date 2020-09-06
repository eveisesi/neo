// Credit: Anderson Sipe @RedVentures

package neo

import (
	"time"
)

type Modifier interface{}

type ModValue interface{}

type ColVal struct {
	Column string
	Value  ModValue
}

type ColValIn struct {
	Column string
	Values []ModValue
}

type ColValOr struct {
	Values []Modifier
}

type ColValStr struct {
	Column string
	Value  string
}

type ColValTime struct {
	Column string
	Value  time.Time
}

type ColValBool struct {
	Column string
	Value  bool
}

type ColValUint struct {
	Column string
	Value  uint
}

type ColValUint64 struct {
	Column string
	Value  uint64
}

type ColValSlcStr struct {
	Column string
	Value  []string
}

type ColValSlcUint64 struct {
	Column string
	Value  []uint64
}

type ColValSlcUint struct {
	Column string
	Value  []uint
}

// This are the new Modifiers for Mongo
type EqualTo ColVal
type NotEqualTo ColVal
type GreaterThan ColVal
type GreaterThanEqualTo ColVal
type LessThan ColVal
type LessThanEqualTo ColVal
type NotEqual ColVal
type In ColValIn
type NotIn ColValIn

// Legacy Modifiers, will most likely be deprecated after we move off of MySQL
// should Mongo perform
type EqualToStr ColValStr
type NotEqualToStr ColValStr
type EqualToTime ColValTime
type NotEqualToTime ColValTime
type GreaterThanTime ColValTime
type LessThanTime ColValTime
type EqualToBool ColValBool
type NotEqualToBool ColValBool
type GreaterThanUint64 ColValUint64
type GreaterThanEqualToUint64 ColValUint64
type LessThanUint64 ColValUint64
type LessThanEqualToUint64 ColValUint64
type EqualToUint64 ColValUint64
type NotEqualToUint64 ColValUint64
type GreaterThanUint ColValUint
type GreaterThanEqualToUint ColValUint
type LessThanUint ColValUint
type LessThanEqualToUint ColValUint
type EqualToUint ColValUint
type NotEqualToUint ColValUint
type InUint ColValSlcUint
type InUint64 ColValSlcUint64
type LimitModifier int

type OrderModifier struct {
	Column string
	Sort   Sort
}

type Sort string

const (
	SortAsc  Sort = "ASC"
	SortDesc Sort = "DESC"
)

var AllSort = []Sort{
	SortAsc,
	SortDesc,
}

func (e Sort) IsValid() bool {
	switch e {
	case SortAsc, SortDesc:
		return true
	}
	return false
}

func (e Sort) String() string {
	return string(e)
}
