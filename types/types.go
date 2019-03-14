package types

import (
	//"gopkg.in/mgo.v2/bson"
	"github.com/globalsign/mgo/bson"

	"bytes"
	"compress/gzip"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
)

// ** fork from https://github.com/jmoiron/sqlx/tree/master/types

// GzippedText is a []byte which transparently gzips data being submitted to
// a database and ungzips data being Scanned from a database.
type GzippedText []byte

// JSONText is a json.RawMessage, which is a []byte underneath.
// Value() validates the json format in the source, and returns an error if
// the json is not valid.  Scan does no validation.  JSONText additionally
// implements `Unmarshal`, which unmarshals the json within to an interface{}
type JSONText json.RawMessage

// BitBool is an implementation of a bool for the MySQL type BIT(1).
// This type allows you to avoid wasting an entire byte for MySQL's boolean type TINYINT.
type BitBool bool

// SliceStr string
type SliceStr []string

// SliceInf interface{}
type SliceInf []interface{}

// JSONMap 任意对象
type JSONMap map[string]interface{}

// BigInt 大整型
type BigInt struct {
	big.Int
}

// Value implements the driver.Valuer interface, gzipping the raw value of
// this GzippedText.
func (g GzippedText) Value() (driver.Value, error) {
	b := make([]byte, 0, len(g))
	buf := bytes.NewBuffer(b)
	w := gzip.NewWriter(buf)
	w.Write(g)
	w.Close()
	return buf.Bytes(), nil

}

// Scan implements the sql.Scanner interface, ungzipping the value coming off
// the wire and storing the raw result in the GzippedText.
func (g *GzippedText) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for GzippedText")
	}
	reader, err := gzip.NewReader(bytes.NewReader(source))
	if err != nil {
		return err
	}
	defer reader.Close()
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	*g = GzippedText(b)
	return nil
}

// MarshalJSON returns j as the JSON encoding of j.
func (j JSONText) MarshalJSON() ([]byte, error) {
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data
func (j *JSONText) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONText: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil

}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j JSONText) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`{}`), nil // tudy: default value
	}
	var m json.RawMessage
	var err = j.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	return []byte(j), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *JSONText) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for JSONText")
	}
	*j = JSONText(append((*j)[0:0], source...))
	return nil
}

// Unmarshal unmarshal's the json in j to v, as in json.Unmarshal.
func (j *JSONText) Unmarshal(v interface{}) error {
	return json.Unmarshal([]byte(*j), v)
}

// Pretty printing for JSONText types
func (j JSONText) String() string {
	return string(j)
}

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

// Value of SliceStr
func (d SliceStr) Value() (driver.Value, error) {
	var (
		b   []byte
		err error
	)
	if d == nil || len(d) == 0 {
		return []byte("[]"), nil
	}
	if b, err = json.Marshal(d); err != nil {
		return nil, err
	}
	return b, err
}

// Scan of SliceStr
func (d *SliceStr) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for SliceStr")
	}
	return json.Unmarshal(source, d)
}

// Value of JSONMap
func (d JSONMap) Value() (driver.Value, error) {
	var (
		b   []byte
		err error
	)
	if d == nil {
		b = []byte(`{}`)
		return b, err
	}
	if b, err = json.Marshal(d); err != nil {
		return nil, err
	}
	return b, err
}

// Value of SliceInf
func (d SliceInf) Value() (driver.Value, error) {
	var (
		b   []byte
		err error
	)
	if d == nil || len(d) == 0 {
		return []byte("[]"), nil
	}
	if b, err = json.Marshal(d); err != nil {
		return nil, err
	}
	return b, err
}

// Scan of SliceInf
func (d *SliceInf) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for SliceInf")
	}
	return json.Unmarshal(source, d)
}

// Scan of JSONMap
func (d *JSONMap) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for JSONMap")
	}
	return json.Unmarshal(source, d)
}

// Scan of BigInt
func (d *BigInt) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		d.SetString(string(v), 10)
	case int:
		d.SetInt64(int64(v))
	case int64:
		d.SetInt64(v)
	default:
	}
	return nil
}

// Value of BigInt
func (d BigInt) Value() (driver.Value, error) {
	return d.Int.String(), nil
}

// GetBSON for mgo
func (d *BigInt) GetBSON() (interface{}, error) {
	v, err := bson.ParseDecimal128(d.String())
	if err != nil {
		return nil, err
	}
	return v, nil
}

// SetBSON for mgo
func (d *BigInt) SetBSON(raw bson.Raw) error {
	switch raw.Kind {
	case bson.ElementDecimal128:
		if bsonData, err := hex.DecodeString(fmt.Sprintf(`18000000136400%x00`, raw.Data)); err == nil {
			var bsonValue struct{ D interface{} }
			if err := bson.Unmarshal(bsonData, &bsonValue); err == nil {
				if dec128, ok := bsonValue.D.(bson.Decimal128); ok {
					d.SetString(dec128.String(), 10)
				}
			} else {
				return err
			}
		} else {
			return err
		}
	default:
		d.SetInt64(0)
	}
	return nil
}

//
func init() {
	var err error

	// accession and uid
	if err = initIds(); err != nil {
		panic(err)
	}
}
