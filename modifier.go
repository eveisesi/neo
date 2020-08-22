// Credit: Anderson Sipe @RedVentures

package neo

import "time"

type Modifier interface{}

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

type EqualToStr ColValStr

// func (r EqualToStr) Modify() ModValue {
// 	return fmt.Sprintf("%s = %s", r.Column, r.Value)
// }

type NotEqualToStr ColValStr

// func (r NotEqualToStr) Modify() ModValue {
// 	return fmt.Sprintf("%s != %s", r.Column, r.Value)
// }

type EqualToTime ColValTime

type NotEqualToTime ColValTime

type GreaterThanTime ColValTime

type LessThanTime ColValTime

type EqualToBool ColValBool

// func (r EqualToBool) Modify() ModValue {
// 	return fmt.Sprintf("%s = %t", r.Column, r.Value)
// }

type NotEqualToBool ColValBool

// func (r NotEqualToBool) Modify() ModValue {
// 	return fmt.Sprintf("%s != %t", r.Column, r.Value)
// }

type GreaterThanUint64 ColValUint64

// func (r GreaterThanUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s > %d", r.Column, r.Value)
// }

type GreaterThanEqualToUint64 ColValUint64

// func (r GreaterThanEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s >= %d", r.Column, r.Value)
// }

type LessThanUint64 ColValUint64

// func (r LessThanUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s < %d", r.Column, r.Value)
// }

type LessThanEqualToUint64 ColValUint64

// func (r LessThanEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s <= %d", r.Column, r.Value)
// }

type EqualToUint64 ColValUint64

// func (r EqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s = %d", r.Column, r.Value)
// }

type NotEqualToUint64 ColValUint64

// func (r NotEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type GreaterThanUint ColValUint

// func (r GreaterThanUint) Modify() ModValue {
// 	return fmt.Sprintf("%s > %d", r.Column, r.Value)
// }

type GreaterThanEqualToUint ColValUint

// func (r GreaterThanEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s >= %d", r.Column, r.Value)
// }

type LessThanUint ColValUint

// func (r LessThanUint) Modify() ModValue {
// 	return fmt.Sprintf("%s < %d", r.Column, r.Value)
// }

type LessThanEqualToUint ColValUint

// func (r LessThanEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s <= %d", r.Column, r.Value)
// }

type EqualToUint ColValUint

// func (r EqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s = %d", r.Column, r.Value)
// }

type NotEqualToUint ColValUint

// func (r NotEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type InUint ColValSlcUint
type InUint64 ColValSlcUint64

// func (r InUint) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type LimitModifier int

// func (r LimitModifier) Modify() ModValue {
// 	return int(r)
// }

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
