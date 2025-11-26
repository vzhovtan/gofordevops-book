package workflow_test

import (
	"bytes"
	"fmt"
	"testing"
	"workflow"
)

var tRoot = "./testdata"
var tStat = []string{"testdata/gopher1.png", "testdata/gopher2.png", "testdata/gopher3.png", "testdata/gopher4.png"}

func TestGetStat(t *testing.T) {
	out := new(bytes.Buffer)
	workflow.GetStat(tRoot, out)
	if out.String() != fmt.Sprintf("%s\n", tStat) {
		t.Errorf("test GetStat failed - results not match\nGot:\n%v\nExpected:\n%v", out.String(), tStat)
	}
}
