package mysql

import (
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

// BuildQueryModifiers returns a Slice of SQL Boiler's QueryMods that can be used in to dynamically build a query
func BuildQueryModifiers(tableName string, modifiers ...neo.Modifier) []qm.QueryMod {

	var mods = make([]qm.QueryMod, 0)
	for _, a := range modifiers {
		switch o := a.(type) {
		case neo.EqualToStr:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s = ?", tableName, o.Column), o.Value))
		case neo.NotEqualToStr:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s != ?", tableName, o.Column), o.Value))
		case neo.EqualToBool:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s = ?", tableName, o.Column), o.Value))
		case neo.NotEqualToBool:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s != ?", tableName, o.Column), o.Value))
		case neo.EqualToUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s = ?", tableName, o.Column), o.Value))
		case neo.NotEqualToUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s != ?", tableName, o.Column), o.Value))
		case neo.GreaterThanEqualToUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s >= ?", tableName, o.Column), o.Value))
		case neo.GreaterThanUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s > ?", tableName, o.Column), o.Value))
		case neo.LessThanUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s < ?", tableName, o.Column), o.Value))
		case neo.LessThanEqualToUint64:
			mods = append(mods, qm.Where(fmt.Sprintf("%s.%s <= ?", tableName, o.Column), o.Value))
		case neo.InUint64:
			mods = append(mods, qm.WhereIn(fmt.Sprintf("%s.%s IN ?", tableName, o.Column), SliceUint64ToSliceInterface(o.Value)...))
		case neo.LimitModifier:
			mods = append(mods, qm.Limit(int(o)))
		case neo.OrderModifier:
			mods = append(mods, qm.OrderBy(fmt.Sprintf("%s.%s %s", tableName, o.Column, o.Sort.String())))
		case neo.EqualToTime:
			mods = append(mods, qm.Where(fmt.Sprintf("%s = ?", o.Column), o.Value.Format(time.RFC3339)))
		case neo.NotEqualToTime:
			mods = append(mods, qm.Where(fmt.Sprintf("%s != ?", o.Column), o.Value.Format(time.RFC3339)))
		case neo.GreaterThanTime:
			mods = append(mods, qm.Where(fmt.Sprintf("%s > ?", o.Column), o.Value.Format(time.RFC3339)))
		case neo.LessThanTime:
			mods = append(mods, qm.Where(fmt.Sprintf("%s < ?", o.Column), o.Value.Format(time.RFC3339)))
		}

	}

	return mods

}

// BuildModifiers returns a slice of string that can still be used to dynamically build a where statement.
// The benefit here is that you don't have to do any blank magic that I still haven't figured out to get this
// to render just there where clause....ask me about this later.
func BuildJoinCondition(tableName string, modifiers ...neo.Modifier) ([]string, []interface{}) {

	var mods = make([]string, 0)
	var args = make([]interface{}, 0)
	for _, a := range modifiers {
		switch o := a.(type) {
		case neo.EqualToStr:
			mods = append(mods, fmt.Sprintf("%s.%s = ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.NotEqualToStr:
			mods = append(mods, fmt.Sprintf("%s.%s != ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.EqualToBool:
			mods = append(mods, fmt.Sprintf("%s.%s = ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.NotEqualToBool:
			mods = append(mods, fmt.Sprintf("%s.%s != ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.EqualToUint64:
			mods = append(mods, fmt.Sprintf("%s.%s = ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.NotEqualToUint64:
			mods = append(mods, fmt.Sprintf("%s.%s != ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.GreaterThanEqualToUint64:
			mods = append(mods, fmt.Sprintf("%s.%s >= ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.GreaterThanUint64:
			mods = append(mods, fmt.Sprintf("%s.%s > ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.LessThanUint64:
			mods = append(mods, fmt.Sprintf("%s.%s < ?", tableName, o.Column))
			args = append(args, o.Value)
		case neo.LessThanEqualToUint64:
			mods = append(mods, fmt.Sprintf("%s.%s <= ?", tableName, o.Column))
			args = append(args, o.Value)
			// case neo.InUint64:
			// 	mods = append(mods, qm.WhereIn(fmt.Sprintf("%s IN ?", o.Column), SliceUint64ToSliceInterface(o.Value.([]uint64))...))
			// case neo.LimitModifier:
			// 	mods = append(mods, qm.Limit(int(o)))
			// case neo.OrderModifier:
			// 	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", o.Column, o.Sort.String())))
		}

	}

	return mods, args

}

func SliceUint64ToSliceInterface(a []uint64) []interface{} {
	s := make([]interface{}, len(a))
	for i, v := range a {
		s[i] = v
	}
	return s
}
