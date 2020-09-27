package resolvers

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/eveisesi/neo"
)

func buildOperators(filter interface{}) (mods []*neo.Operator, err error) {

	v := reflect.Indirect(reflect.ValueOf(filter))
	if !v.IsValid() {
		return
	}

	if v.Kind() != reflect.Struct {
		err = errors.New(`underlying filter type must be "struct"`)
		return
	}

	for i := 0; i < v.NumField(); i++ {

		if v.Field(i).IsNil() {
			continue
		}
		// The name of the struct field. i.e. "SolarSystemID", "WarID", etc
		colName := strcase.ToLowerCamel(v.Type().Field(i).Name)
		// The Name of the Filter Type i.e. "IntFilter", "BooleanFilter", etc
		fType := reflect.Indirect(v.Field(i)).Type().Name()
		// The value of our filter type
		fieldValue := reflect.Indirect(v.Field(i))

		// If colName points to a nested struct field, mutate the name so that this will work down stream
		if strings.HasPrefix(colName, "attackers") {
			colName = fmt.Sprintf("attackers.%s", strcase.ToLowerCamel(colName[9:]))
		} else if strings.HasPrefix(colName, "victim") {
			colName = fmt.Sprintf("victim.%s", strcase.ToLowerCamel(colName[6:]))
		}

		switch fieldValue.Kind() {
		case reflect.Int:
			if mod := getSimpleOperator(colName, reflect.Indirect(v.Field(i))); mod != nil {
				mods = append(mods, mod)
			}

			continue
		default:
			// let it through
		}

		// Check if we have a more concrete filter type. I.E. ColumnSort only
		// has two defined fields
		// if fType == "ColumnSort" {
		// 	mods = append(mods, getSortOperator(fieldValue))
		// 	continue
		// }

		// We didn't hit any of the prior checks so assume we have a more
		// complex nested struct. Loop through the nexted struct fields
		// containing our operator types.
		// E.G. IntFilter struct {EQ: 1, LT: 2}
		for j := 0; j < fieldValue.NumField(); j++ {

			if fieldValue.Field(j).IsNil() {
				continue
			}

			// Name of the comparison operator. E.G. "Eq", "Gt", "Contains", etc
			n := fieldValue.Type().Field(j).Name

			// The indirect value of our field pointers. I.E. *int = 1
			iv := reflect.Indirect(fieldValue.Field(j))

			switch fType {
			// case "StringFilterInput":
			// 	if mod := getStrOperator(colName, n, iv); mod != nil {
			// 		mods = append(mods, mod)
			// 	}
			case "IntFilterInput":
				if mod := getIntOperator(colName, n, iv); mod != nil {
					mods = append(mods, mod)
				}
				continue
			case "BooleanFilterInput":
				if mod := getBoolOperator(colName, n, iv); mod != nil {
					mods = append(mods, mod)
				}
			default:
				panic("unsupported filter type")
			}
		}

	}

	return

}

func getSimpleOperator(col string, v reflect.Value) *neo.Operator {
	switch col {
	case "LimitFilter":
		return neo.NewLimitOperator(v.Int())
	}
	return nil
}

// Determine the string modifier
// func getStrOperator(col, op string, v reflect.Value) neo.Operator {
// 	switch op {
// 	case "Ne": // Not Equal To
// 		return neo.NotEqualTo{
// 			Column: col,
// 			Value:  v.String(),
// 		}
// 	case "Eq": // Equal To
// 		return neo.EqualTo{
// 			Column: col,
// 			Value:  v.String(),
// 		}
// 	}

// 	return nil
// }

// Determine the boolean modifier
func getBoolOperator(col, op string, v reflect.Value) *neo.Operator {
	switch op {
	case "Eq": // Equal To
		return neo.NewEqualOperator(col, v.Bool())
	case "Ne":
		return neo.NewNotEqualOperator(col, v.Bool())
	}
	return nil
}

// Determine the int modifier
func getIntOperator(col, op string, v reflect.Value) *neo.Operator {
	switch op {
	case "Ne": // Not Equal To
		return neo.NewNotEqualOperator(col, v.Int())
	case "Eq": // Equal To
		return neo.NewEqualOperator(col, v.Int())
	case "Gte":
		return neo.NewGreaterThanEqualToOperator(col, v.Int())
	case "Gt": // Greater Than
		return neo.NewGreaterThanOperator(col, v.Int())
	case "Lt": // Less Than
		return neo.NewLessThanOperator(col, v.Int())
	case "Lte":
		return neo.NewLessThanEqualToOperator(col, v.Int())
	case "In":
		return neo.NewInOperator(col, v.Interface())
	case "NotIn":
		return neo.NewNotInOperator(col, v.Interface())
	}

	return nil
}
