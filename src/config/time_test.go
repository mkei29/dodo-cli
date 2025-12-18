package config

import (
	"testing"
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
