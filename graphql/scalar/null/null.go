package null

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/volatiletech/null"
)

func MarshalBool(nb null.Bool) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !nb.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, strconv.FormatBool(nb.Bool))
	})
}

func UnmarshalBool(i interface{}) (null.Bool, error) {

	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewBool(false, false), nil
		}

		b, e := strconv.ParseBool(v)
		if e != nil {
			return null.NewBool(false, false), e
		}

		return null.NewBool(b, true), nil
	case bool:
		return null.NewBool(v, true), nil

	default:
		return null.NewBool(false, false), fmt.Errorf("%v is not a valid bool", v)
	}
}

func MarshalFloat64(nf null.Float64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !nf.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, strconv.FormatFloat(nf.Float64, 'f', -1, 64))
	})
}

func UnmarshalFloat64(i interface{}) (null.Float64, error) {
	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewFloat64(float64(0), false), nil
		}

		f, e := strconv.ParseFloat(v, 64)
		if e != nil {
			return null.NewFloat64(float64(0), false), e
		}

		return null.NewFloat64(f, true), nil
	case float32:
		return null.NewFloat64(float64(v), true), nil
	case float64:
		return null.NewFloat64(v, true), nil
	default:
		return null.NewFloat64(float64(0), false), fmt.Errorf("%v is not a valid float64", v)
	}
}

func MarshalInt64(ni null.Int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !ni.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, strconv.FormatInt(ni.Int64, 10))
	})
}

func UnmarshalInt64(i interface{}) (null.Int64, error) {
	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewInt64(int64(0), false), nil
		}

		f, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			return null.NewInt64(int64(0), false), e
		}

		return null.NewInt64(f, true), nil
	case float32:
		return null.NewInt64(int64(v), true), nil
	case float64:
		return null.NewInt64(int64(v), true), nil
	case int:
		return null.NewInt64(int64(v), true), nil
	case int8:
		return null.NewInt64(int64(v), true), nil
	case int16:
		return null.NewInt64(int64(v), true), nil
	case int32:
		return null.NewInt64(int64(v), true), nil
	case int64:
		return null.NewInt64(v, true), nil
	case uint:
		return null.NewInt64(int64(v), true), nil
	case uint8:
		return null.NewInt64(int64(v), true), nil
	case uint16:
		return null.NewInt64(int64(v), true), nil
	case uint32:
		return null.NewInt64(int64(v), true), nil
	case uint64:
		return null.NewInt64(int64(v), true), nil
	default:
		return null.NewInt64(int64(0), false), fmt.Errorf("%v is not a valid int64", v)
	}
}

func MarshalString(ns null.String) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !ns.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, ns.String)
	})
}

func UnmarshalString(i interface{}) (null.String, error) {
	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewString("", false), nil
		}
		return null.NewString(v, true), nil
	default:
		return null.NewString("", false), fmt.Errorf("%v is not a valid string", v)
	}
}

func MarshalTime(nt null.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !nt.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, nt.Time.Format("2006-01-02 15:03:04"))
	})
}

func UnmarshalTime(i interface{}) (null.Time, error) {

	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewTime(time.Now(), false), nil
		}

		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return null.NewTime(time.Now(), false), fmt.Errorf("%v is not formatted correctly. Please format time according to RFC3339 (%s)", v, time.RFC3339)
		}

		return null.NewTime(parsed, true), nil
	default:
		return null.NewTime(time.Now(), false), fmt.Errorf("%v is not a valid string", v)
	}

}

func MarshalUint64(nu null.Uint64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !nu.Valid {
			io.WriteString(w, `null`)
			return
		}

		io.WriteString(w, strconv.FormatUint(nu.Uint64, 10))
	})
}

func UnmarshalUint64(i interface{}) (null.Uint64, error) {
	switch v := i.(type) {
	case string:
		if v == "null" {
			return null.NewUint64(uint64(0), false), nil
		}

		u, e := strconv.ParseUint(v, 10, 64)
		if e != nil {
			return null.NewUint64(uint64(0), false), e
		}

		return null.NewUint64(u, true), nil
	case float32:
		return null.NewUint64(uint64(v), true), nil
	case float64:
		return null.NewUint64(uint64(v), true), nil
	case int:
		return null.NewUint64(uint64(v), true), nil
	case int8:
		return null.NewUint64(uint64(v), true), nil
	case int16:
		return null.NewUint64(uint64(v), true), nil
	case int32:
		return null.NewUint64(uint64(v), true), nil
	case int64:
		return null.NewUint64(uint64(v), true), nil
	case uint:
		return null.NewUint64(uint64(v), true), nil
	case uint8:
		return null.NewUint64(uint64(v), true), nil
	case uint16:
		return null.NewUint64(uint64(v), true), nil
	case uint32:
		return null.NewUint64(uint64(v), true), nil
	case uint64:
		return null.NewUint64(v, true), nil
	default:
		return null.NewUint64(uint64(0), false), fmt.Errorf("%v is not a valid int64", v)
	}
}
