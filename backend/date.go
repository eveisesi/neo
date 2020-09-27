package neo

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

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
	case time.Time:
		*d = Date{v}
		return nil
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
