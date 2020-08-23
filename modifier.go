// Credit: Anderson Sipe @RedVentures

package neo

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

type Modifier interface {
	CacheKey() string
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

type EqualToStr ColValStr

func (r EqualToStr) CacheKey() string {
	return fmt.Sprintf("%s:ets:%s", r.Column, r.Value)
}

// func (r EqualToStr) Modify() ModValue {
// 	return fmt.Sprintf("%s = %s", r.Column, r.Value)
// }

type NotEqualToStr ColValStr

func (r NotEqualToStr) CacheKey() string {
	return fmt.Sprintf("%s:nets:%s", r.Column, r.Value)
}

// func (r NotEqualToStr) Modify() ModValue {
// 	return fmt.Sprintf("%s != %s", r.Column, r.Value)
// }

type EqualToTime ColValTime

func (r EqualToTime) CacheKey() string {
	return fmt.Sprintf("%s:ett:%s", r.Column, r.Value.Format("2006-01-02-15-04-05"))
}

type NotEqualToTime ColValTime

func (r NotEqualToTime) CacheKey() string {
	return fmt.Sprintf("%s:nett:%s", r.Column, r.Value.Format("2006-01-02-15-04-05"))
}

type GreaterThanTime ColValTime

func (r GreaterThanTime) CacheKey() string {
	return fmt.Sprintf("%s:gtt:%s", r.Column, r.Value.Format("2006-01-02-15-04-05"))
}

type LessThanTime ColValTime

func (r LessThanTime) CacheKey() string {
	return fmt.Sprintf("%s:ltt:%s", r.Column, r.Value.Format("2006-01-02-15-04-05"))
}

type EqualToBool ColValBool

func (r EqualToBool) CacheKey() string {
	return fmt.Sprintf("%s:etb:%t", r.Column, r.Value)
}

// func (r EqualToBool) Modify() ModValue {
// 	return fmt.Sprintf("%s = %t", r.Column, r.Value)
// }

type NotEqualToBool ColValBool

func (r NotEqualToBool) CacheKey() string {
	return fmt.Sprintf("%s:netb:%t", r.Column, r.Value)
}

// func (r NotEqualToBool) Modify() ModValue {
// 	return fmt.Sprintf("%s != %t", r.Column, r.Value)
// }

type GreaterThanUint64 ColValUint64

func (r GreaterThanUint64) CacheKey() string {
	return fmt.Sprintf("%s:gtu64:%d", r.Column, r.Value)
}

// func (r GreaterThanUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s > %d", r.Column, r.Value)
// }

type GreaterThanEqualToUint64 ColValUint64

func (r GreaterThanEqualToUint64) CacheKey() string {
	return fmt.Sprintf("%s:gteu64:%d", r.Column, r.Value)
}

// func (r GreaterThanEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s >= %d", r.Column, r.Value)
// }

type LessThanUint64 ColValUint64

func (r LessThanUint64) CacheKey() string {
	return fmt.Sprintf("%s:ltu64:%d", r.Column, r.Value)
}

// func (r LessThanUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s < %d", r.Column, r.Value)
// }

type LessThanEqualToUint64 ColValUint64

func (r LessThanEqualToUint64) CacheKey() string {
	return fmt.Sprintf("%s:lteu64:%d", r.Column, r.Value)
}

// func (r LessThanEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s <= %d", r.Column, r.Value)
// }

type EqualToUint64 ColValUint64

func (r EqualToUint64) CacheKey() string {
	return fmt.Sprintf("%s:eu64:%d", r.Column, r.Value)
}

// func (r EqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s = %d", r.Column, r.Value)
// }

type NotEqualToUint64 ColValUint64

func (r NotEqualToUint64) CacheKey() string {
	return fmt.Sprintf("%s:neu64:%d", r.Column, r.Value)
}

// func (r NotEqualToUint64) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type GreaterThanUint ColValUint

func (r GreaterThanUint) CacheKey() string {
	return fmt.Sprintf("%s:gui:%d", r.Column, r.Value)
}

// func (r GreaterThanUint) Modify() ModValue {
// 	return fmt.Sprintf("%s > %d", r.Column, r.Value)
// }

type GreaterThanEqualToUint ColValUint

func (r GreaterThanEqualToUint) CacheKey() string {
	return fmt.Sprintf("%s:gteui:%d", r.Column, r.Value)
}

// func (r GreaterThanEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s >= %d", r.Column, r.Value)
// }

type LessThanUint ColValUint

func (r LessThanUint) CacheKey() string {
	return fmt.Sprintf("%s:ltui:%d", r.Column, r.Value)
}

// func (r LessThanUint) Modify() ModValue {
// 	return fmt.Sprintf("%s < %d", r.Column, r.Value)
// }

type LessThanEqualToUint ColValUint

func (r LessThanEqualToUint) CacheKey() string {
	return fmt.Sprintf("%s:lteui:%d", r.Column, r.Value)
}

// func (r LessThanEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s <= %d", r.Column, r.Value)
// }

type EqualToUint ColValUint

func (r EqualToUint) CacheKey() string {
	return fmt.Sprintf("%s:eui:%d", r.Column, r.Value)
}

// func (r EqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s = %d", r.Column, r.Value)
// }

type NotEqualToUint ColValUint

func (r NotEqualToUint) CacheKey() string {
	return fmt.Sprintf("%s:neui:%d", r.Column, r.Value)
}

// func (r NotEqualToUint) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type InUint ColValSlcUint

func (r InUint) CacheKey() string {

	pieces := []string{}
	for _, value := range r.Value {
		pieces = append(pieces, fmt.Sprintf("%d", value))
	}

	h := sha1.New()
	h.Write([]byte(strings.Join(pieces, "-")))
	return fmt.Sprintf("%x", h.Sum(nil))

}

type InUint64 ColValSlcUint64

func (r InUint64) CacheKey() string {

	pieces := []string{}
	for _, value := range r.Value {
		pieces = append(pieces, fmt.Sprintf("%d", value))
	}

	h := sha1.New()
	h.Write([]byte(strings.Join(pieces, "-")))
	return fmt.Sprintf("%x", h.Sum(nil))

}

// func (r InUint) Modify() ModValue {
// 	return fmt.Sprintf("%s != %d", r.Column, r.Value)
// }

type LimitModifier int

func (r LimitModifier) CacheKey() string {
	return fmt.Sprintf("limit:%d", int(r))
}

// func (r LimitModifier) Modify() ModValue {
// 	return int(r)
// }

type OrderModifier struct {
	Column string
	Sort   Sort
}

func (r OrderModifier) CacheKey() string {
	return fmt.Sprintf("order:%s:%s", r.Column, r.Sort.String())
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
