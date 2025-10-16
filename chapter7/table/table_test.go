package table_test

import (
	"fmt"
	"table"
	"testing"
)

func TestAdd1(t *testing.T) {
	var tests = []struct {
		a, b int
		want int
	}{
		{0, 1, 0},
		{1, 0, 0},
		{2, -2, -4},
		{1, -1, -1},
		{10, 100, 1000},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("Ruuning Mutiply function with the arguments %d and %d", tt.a, tt.b)
		t.Run(testname, func(t *testing.T) {
			got := table.Multiply(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("The test is failing, got %d, want %d", got, tt.want)
			}
		})
	}
}
