package flags_test

import (
	"bytes"
	"flags"
	"testing"
)

const testFullResult = `├───project
│	├───file.txt (232b)
│	└───gopher.png (70372b)
├───static
│	├───a_lorem
│	│	├───gopher.png (70372b)
│	│	└───ipsum
│	│		└───gopher.png (70372b)
│	├───css
│	│	└───body.css (28b)
│	├───html
│	│	└───index.html (57b)
│	├───js
│	│	└───site.js (10b)
│	└───z_lorem
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
├───zline
│	└───lorem
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
└───zzfile.txt (empty)
`

func TestBuildDirTree(t *testing.T) {
	out := new(bytes.Buffer)
	x := true
	fflag := &x
	var last bool
	var lastlist []bool
	err := flags.BuildDirTree(out, "testdata", last, fflag, lastlist)
	if err != nil {
		t.Errorf("test TreeFull Failed - error")
	}
	result := out.String()
	if result != testFullResult {
		t.Errorf("test TreeFull Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testFullResult)
	}
}

const testDirResult = `├───project
├───static
│	├───a_lorem
│	│	└───ipsum
│	├───css
│	├───html
│	├───js
│	└───z_lorem
│		└───ipsum
└───zline
	└───lorem
		└───ipsum
`

func TestTreeDir(t *testing.T) {
	out := new(bytes.Buffer)
	x := false
	fflag := &x
	var last bool
	var lastlist []bool
	err := flags.BuildDirTree(out, "testdata", last, fflag, lastlist)
	if err != nil {
		t.Errorf("test TreeDir Failed - error")
	}
	result := out.String()
	if result != testDirResult {
		t.Errorf("test TreeDir Failed - results not match\nGot:\n%v\nExpected:\n%v\n", result, testDirResult)
	}
}
