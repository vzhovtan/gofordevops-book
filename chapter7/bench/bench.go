package bench

func FindLarger(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func CPUConsumer() {
	limit := 5_000_000_000
	result := 0.0
	for i := 0; i < limit; i++ {
		result += float64(i) * float64(i)
	}
}

func MemoryConsumer(sizeInMB int) []byte {
	data := make([]byte, sizeInMB*1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	return data
}
