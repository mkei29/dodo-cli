package config

import (
	"fmt"
	"time"
)

// SerializableTime is a custom type to process time in json and yaml format.
type SerializableTime struct {
	time.Time
}

func NewSerializableTime(s string) (SerializableTime, error) {
	// If the string is empty, return an empty time.
	if s == "" {
		return SerializableTime{}, nil
	}

	// In other cases, parse the string into a time.
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return SerializableTime{}, fmt.Errorf("failed to unmarshal time: %w", err)
	}
	return SerializableTime{t}, nil
}

func NewSerializableTimeFromTime(t time.Time) SerializableTime {
	return SerializableTime{t}
}

// String converts the time into a string.
func (t SerializableTime) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func (t *SerializableTime) HasValue() bool {
	return !t.IsZero()
}
