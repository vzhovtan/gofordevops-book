package main

import (
	"cli"
	"flag"
)

func main() {
	fname := flag.String("fileName", "", "HTML file to parse")
	flag.Parse()
	cli.Run(*fname)
}
