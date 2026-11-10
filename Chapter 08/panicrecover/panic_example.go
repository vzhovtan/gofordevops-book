package panicrecover

import (
	"fmt"
	"os"
)

func FileOpenRead(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	content := string(data)
	fmt.Println(content)
}
