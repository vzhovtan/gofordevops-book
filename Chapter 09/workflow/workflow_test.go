package workflow_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"workflow"
)

var tRoot = "./testdata"
var path1 = filepath.Join("testdata", "gopher1.png")
var path2 = filepath.Join("testdata", "gopher2.png")
var path3 = filepath.Join("testdata", "gopher3.png")
var path4 = filepath.Join("testdata", "gopher4.png")
var tStat = []string{path1, path2, path3, path4}

func TestGetStat(t *testing.T) {
	out := new(bytes.Buffer)
	workflow.GetStat(tRoot, out)
	if out.String() != fmt.Sprintf("%s\n", tStat) {
		t.Errorf("test GetStat failed - results not match\nGot:\n%v\nExpected:\n%v", out.String(), tStat)
	}
}
