package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Source: https://github.com/jmoiron/sqlx/blob/master/types/types.go
// 
// For HEXYA: In db.sanitizeQuery() function, call to sqlx.Rebind(sqlx.BindType(db.DriverName()), q) 
// causes invalid SQL statement with json.RawMessage type. So, I change JSONText to string type

//type JSONText json.RawMessage
type JSONText string

var emptyJSON = JSONText("{}")

//// MarshalJSON returns the *j as the JSON encoding of j.
//func (j JSONText) MarshalJSON() ([]byte, error) {
//	if len(j) == 0 {
//		return emptyJSON, nil
//	}
//	return j, nil
//}

//// UnmarshalJSON sets *j to a copy of data
//func (j *JSONText) UnmarshalJSON(data []byte) error {
//	if j == nil {
//		return errors.New("JSONText: UnmarshalJSON on nil pointer")
//	}
//	*j = append((*j)[0:0], data...)
//	return nil
//}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j JSONText) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = j.Unmarshal(&m)
	if err != nil {
		return "{}", err
	}
	return string(j), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *JSONText) Scan(src interface{}) error {
	if src == nil {
		*j = emptyJSON
	} else {
		*j = JSONText(string(src.([]uint8)))
	}
	return nil
}

// Unmarshal unmarshal's the json in j to v, as in json.Unmarshal.
func (j *JSONText) Unmarshal(v interface{}) error {
	if len(*j) == 0 {
		*j = emptyJSON
	}
	return json.Unmarshal([]byte(*j), v)
}

// String supports pretty printing for JSONText types.
func (j JSONText) String() string {
	return string(j)
}

// NullJSONText represents a JSONText that may be null.
// NullJSONText implements the scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullJSONText struct {
	JSONText
	Valid bool // Valid is true if JSONText is not NULL
}

// Scan implements the Scanner interface.
func (n *NullJSONText) Scan(value interface{}) error {
	if value == nil {
		n.JSONText, n.Valid = emptyJSON, false
		return nil
	}
	n.Valid = true
	return n.JSONText.Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullJSONText) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.JSONText.Value()
}

// BitBool is an implementation of a bool for the MySQL type BIT(1).
// This type allows you to avoid wasting an entire byte for MySQL's boolean type TINYINT.
type BitBool bool

// Value implements the driver.Valuer interface,
// and turns the BitBool into a bitfield (BIT(1)) for MySQL storage.
func (b BitBool) Value() (driver.Value, error) {
	if b {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}

// Scan implements the sql.Scanner interface,
// and turns the bitfield incoming from MySQL into a BitBool
func (b *BitBool) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	*b = v[0] == 1
	return nil
}
