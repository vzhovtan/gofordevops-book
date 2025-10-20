package panicrecover

import (
	"fmt"
	"os"
)

func FileOpenRead(fpath string) {
	// Read the file
	data, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	content := string(data)
	fmt.Println(content)
}
