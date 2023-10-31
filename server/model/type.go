package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type ContentOther struct {
	CidNext uint `json:"cid_next"`
}

func (j *ContentOther) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := ContentOther{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}
func (j ContentOther) Value() (driver.Value, error) {
	b, err := json.Marshal(&j)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	return b, nil
}

type ContentFocus struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

func (j *ContentFocus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := ContentFocus{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}
func (j ContentFocus) Value() (driver.Value, error) {
	b, err := json.Marshal(&j)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	return b, nil
}

type Config struct {
	ContactId string `json:"contactId" form:"contactId"`
}

func (j *Config) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := Config{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}
func (j Config) Value() (driver.Value, error) {
	b, err := json.Marshal(&j)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	return b, nil
}

type FileOther struct {
	Width  int `json:"w"`
	Height int `json:"h"`
}

func (j *FileOther) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := FileOther{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}
func (j FileOther) Value() (driver.Value, error) {
	b, err := json.Marshal(&j)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	return b, nil
}
