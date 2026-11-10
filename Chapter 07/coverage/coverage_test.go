package coverage_test

import (
	"coverage"
	"testing"
)

func TestAdd(t *testing.T) {
	got := coverage.Add(2, 3)
	want := 5

	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}

func TestSub(t *testing.T) {
	got := coverage.Sub(10, 5)
	want := 5

	if got != want {
		t.Errorf("Sub(2, 3) = %d; want %d", got, want)
	}
}

func TestMul(t *testing.T) {
	got := coverage.Mul(2, 5)
	want := 10

	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}
