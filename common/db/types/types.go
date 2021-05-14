package types

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"strings"
)

var ErrInvalidType = errors.New("Invalid input type")

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return strings.Join(s, "-/-"), nil
}

func (s *StringArray) Scan(input interface{}) error {
	var str string
	var ok bool
	switch input.(type) {
	case string:
		{
			str, ok = input.(string)
			if !ok {
				return ErrInvalidType
			}
		}
	case []byte:
		{
			bs, ok := input.([]byte)
			str = string(bs)
			if !ok {
				return ErrInvalidType
			}
		}
	}

	if str != "" {
		*s = strings.Split(str, "-/-")
	} else {
		*s = []string{}
	}

	return nil
}

type UintArray []uint

func (u UintArray) Value() (driver.Value, error) {
	r := make([]byte, 8*len(u))
	for i, v := range u {
		binary.LittleEndian.PutUint64(r[8*i:], uint64(v))
	}

	return r, nil
}

func (u *UintArray) Scan(input interface{}) error {
	v, ok := input.([]byte)
	if !ok {
		return ErrInvalidType
	}

	for len(v) > 0 {
		vv := binary.LittleEndian.Uint64(v)
		*u = append(*u, uint(vv))
		v = v[8:]
	}

	return nil
}
