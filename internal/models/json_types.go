package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONStringSlice []string

func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = JSONStringSlice{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}