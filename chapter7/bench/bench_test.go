package bench_test

import (
	"bench"
	"testing"
)

func BenchmarkFindLarger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench.FindLarger(20, 10)
	}
}

func BenchmarkCPUConsumer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench.CPUConsumer()
	}
}

func BenchmarkMemoryConsumer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bench.MemoryConsumer(10)
	}
}
