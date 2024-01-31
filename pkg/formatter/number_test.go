package formatter

import (
	"fmt"
	"testing"
)

func TestPrettyPrintInt64(t *testing.T) {
	testCases := []struct {
		input    int64
		expected string
	}{
		{input: 0, expected: "0"},
		{input: 123, expected: "123"},
		{input: 1000, expected: "1,000"},
		{input: 123456789, expected: "123,456,789"},
		{input: 9876543210, expected: "9,876,543,210"},
		{input: -123456789, expected: "-123,456,789"},
		{input: -9876543210, expected: "-9,876,543,210"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input=%d expected=%s", tc.input, tc.expected), func(t *testing.T) {
			result := PrettyPrintInt64(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %s, but got: %s", tc.expected, result)
			}
		})
	}
}
