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

type OrMod struct {
	Values []Modifier
}

type AndMod struct {
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

type Exists struct {
	Column string
}

type NotExists struct {
	Column string
}

type LimitModifier int

type OrderModifier struct {
	Column string
	Sort   Sort
}

type Sort int

const (
	SortAsc  Sort = 1
	SortDesc Sort = -1
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

func (e Sort) Value() int {
	return int(e)
}
