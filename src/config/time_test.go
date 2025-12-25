package config

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewSerializableTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"2023-10-10T10:10:10Z", "2023-10-10T10:10:10Z", false},
		{"", "", false},
		{"invalid-time", "", true},
	}

	for _, test := range tests {
		result, err := NewSerializableTime(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("NewSerializableTime(%s) error = %v, expected error = %v", test.input, err, test.hasError)
		}
		if result.String() != test.expected {
			t.Errorf("NewSerializableTime(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestSerializableTimeMarshalJSON(t *testing.T) {
	zero := SerializableTime{}
	zeroBytes, err := json.Marshal(zero)
	if err != nil {
		t.Fatalf("json.Marshal(zero) error = %v", err)
	}
	if string(zeroBytes) != `""` {
		t.Fatalf("json.Marshal(zero) = %s, expected \"\"", zeroBytes)
	}

	nonZero := NewSerializableTimeFromTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	nonZeroBytes, err := json.Marshal(nonZero)
	if err != nil {
		t.Fatalf("json.Marshal(nonZero) error = %v", err)
	}
	if string(nonZeroBytes) != `"2025-01-01T00:00:00Z"` {
		t.Fatalf("json.Marshal(nonZero) = %s, expected RFC3339 string", nonZeroBytes)
	}
}
