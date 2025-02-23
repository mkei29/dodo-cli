package main

import (
	"fmt"
	"time"
)

// Serializable is a custom type for string to process time in json and yaml format.
//
// Reference:
// https://kenzo0107.github.io/2020/05/19/2020-05-20-go-json-time/
// https://pkg.go.dev/gopkg.in/yaml.v2#Unmarshaler
type SerializableTime string

func NewSerializableTime(s string) (SerializableTime, error) {
	st := SerializableTime(s)
	_, err := st.Time()
	if err != nil {
		return SerializableTime(""), fmt.Errorf("failed to unmarshal time: %w", err)
	}
	return st, nil
}

// String converts the unix timestamp into a string.
func (t SerializableTime) String() string {
	return string(t)
}

// Time returns a `time.Time` representation of this value.
func (t SerializableTime) Time() (time.Time, error) {
	tt, err := time.Parse(time.RFC3339, string(t))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
	}
	return tt, nil
}

func (t *SerializableTime) IsNull() bool {
	return *t == ""
}
