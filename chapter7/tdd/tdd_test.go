package tdd_test

import (
	"tdd"

	"testing"
)

// first test case validates the correct result returned by the function
func TestAdd1(t *testing.T) {
	result := tdd.Multiply(2, 3)
	expected := 6

	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}
