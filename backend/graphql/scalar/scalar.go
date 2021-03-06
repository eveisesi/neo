package scalar

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalBool(b bool) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatBool(b))
	})
}

func UnmarshalBool(v interface{}) (bool, error) {
	if b, ok := v.(bool); ok {
		return bool(b), nil
	}

	return false, fmt.Errorf("%v is not a bool", v)
}

func MarshalFloat64(f float64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatFloat(f, 'f', -1, 64))
	})
}

func UnmarshalFloat64(v interface{}) (float64, error) {
	if f, ok := v.(float64); ok {
		return float64(f), nil
	}

	return float64(0), fmt.Errorf("%v is not a float64", v)
}

func MarshalInt64(u int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatInt(u, 10))
	})
}

func UnmarshalInt64(v interface{}) (int64, error) {
	if i, ok := v.(int64); ok {
		return int64(i), nil
	}

	return int64(0), fmt.Errorf("%v is not a int64", v)
}

func MarshalString(s string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, s)
	})
}

func UnmarshalString(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return string(s), nil
	}

	return "", fmt.Errorf("%v is not a string", v)
}

func MarshalTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, t.Format("2006-01-02 15:03:04"))
	})
}

func UnmarshalTime(v interface{}) (time.Time, error) {

	if t, ok := v.(string); ok {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return time.Now(), fmt.Errorf("%v is not formatted correctly. Please format time according to RFC3339 (%s)", v, time.RFC3339)
		}

		return parsed, nil
	}

	return time.Now(), fmt.Errorf("%v is not a time", v)
}

func MarshalUint8(u uint8) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(uint64(u), 10))
	})
}

func UnmarshalUint8(v interface{}) (uint8, error) {
	if i, ok := v.(uint8); ok {
		return uint8(i), nil
	}

	return uint8(0), fmt.Errorf("%v is not a uint64", v)
}

func MarshalUint(u uint) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(uint64(u), 10))
	})
}

func UnmarshalUint(v interface{}) (uint, error) {
	if i, ok := v.(uint); ok {
		return uint(i), nil
	}

	return uint(0), fmt.Errorf("%v is not a uint64", v)
}

func MarshalUint64(u uint64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(u, 10))
	})
}

func UnmarshalUint64(v interface{}) (uint64, error) {
	if i, ok := v.(uint64); ok {
		return uint64(i), nil
	}

	return uint64(0), fmt.Errorf("%v is not a uint64", v)
}
