package stl

import (
	"fmt"
	"strconv"
)

func StrconvExamples() {
	// Example1: String to Integer conversions
	str := "42"
	num, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Atoi(\"%s\") = %d\n", str, num)

	// Example2: ParseInt - parse with base
	str64 := "1010"
	num64, _ := strconv.ParseInt(str64, 2, 64) // binary
	fmt.Printf("ParseInt(\"%s\", base 2) = %d\n", str64, num64)

	// Example3: ParseFloat
	floatStr := "3.14159"
	f, _ := strconv.ParseFloat(floatStr, 64)
	fmt.Printf("ParseFloat(\"%s\") = %f\n", floatStr, f)

	// Example4: ParseBool
	boolStr := "true"
	b, _ := strconv.ParseBool(boolStr)
	fmt.Printf("ParseBool(\"%s\") = %v\n", boolStr, b)

	// Example5: Integer to String conversions
	intVal := 99
	intStr := strconv.Itoa(intVal)
	fmt.Printf("Itoa(%d) = \"%s\"\n", intVal, intStr)

	// Example6: FormatInt with base
	int64Val := int64(255)
	hexStr := strconv.FormatInt(int64Val, 16)
	fmt.Printf("FormatInt(%d, base 16) = \"%s\"\n", int64Val, hexStr)

	// Example7: FormatFloat
	floatVal := 3.14159
	formatted := strconv.FormatFloat(floatVal, 'f', 2, 64)
	fmt.Printf("FormatFloat(%.5f, 'f', 2) = \"%s\"\n", floatVal, formatted)

	// Example8: FormatBool
	boolVal := true
	boolFormatted := strconv.FormatBool(boolVal)
	fmt.Printf("FormatBool(%v) = \"%s\"\n", boolVal, boolFormatted)
}
