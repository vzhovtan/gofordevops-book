package unittest_test

// The file with unit test for the sum function
import (
	"testing"
	"unittest"
)

// The first test case validates the correct result returned by the function
func TestAdd1(t *testing.T) {
	result := unittest.Add(2, 3)
	expected := 5

	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}

// The second test case validates the bad result returned by the function
func TestAdd2(t *testing.T) {
	result := unittest.Add(2, 3)
	expected := 6

	if result == expected {
		t.Errorf("Adding(2, 3) = %d; want %d, validating failing scenario", result, expected)
	}
}
