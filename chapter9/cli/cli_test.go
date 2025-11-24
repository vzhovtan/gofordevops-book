package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var long = `[html]
[html head]
[html head meta]
[html head meta]
[html head meta]
[html head meta]
[html head title]
[html head link]
[html head link]
[html head link]
[html body]
[html body header]
[html body header div]
[html body header div a]
[html body header nav]
[html body header nav a]
[html body header nav a img]
[html body header nav button]
[html body header nav button div]
[html body header nav ul]
[html body header nav ul li]
[html body header nav ul li a]
[html body header nav ul li]
[html body header nav ul li a]
[html body header nav ul li]
[html body header nav ul li a]
[html body header nav ul li]
[html body header nav ul li a]
[html body header nav ul li]
[html body header nav ul li a]
[html body header nav ul li]
[html body header nav ul li a]
[html body main]
[html body main div]
[html body main div div]
[html body main div div]
[html body main div div section]
[html body main div div section h1]
[html body main div div section h1 strong]
[html body main div div section h1 strong]
[html body main div div section h1 strong]
[html body main div div section i]
[html body main div div section a]
[html body main div div section a img]
[html body main div div section p]
[html body main div div section p br]
[html body main div div section]
[html body main div div section div]
[html body main div div section div h2]
[html body main div div section div a]
[html body main div div section div]
[html body main div div section div textarea]
[html body main div div section div]
[html body main div div section div pre]
[html body main div div section div pre noscript]
[html body main div div section div]
[html body main div div section div select]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div select option]
[html body main div div section div div]
[html body main div div section div div button]
[html body main div div section div div div]
[html body main div div section div div div button]
[html body main div div section div div div a]
[html body main div div section]
[html body main div div section h2]
[html body main div div section div]
[html body main div div section div a]
[html body main div div section]
[html body main div div section h2]
[html body main div div section div]
[html body main div div section div iframe]
[html body footer]
[html body footer div]
[html body footer div img]
[html body footer div ul]
[html body footer div ul li]
[html body footer div ul li a]
[html body footer div ul li]
[html body footer div ul li a]
[html body footer div ul li]
[html body footer div ul li a]
[html body footer div ul li]
[html body footer div ul li a]
[html body footer div a]
`

var short = `[html]
[html head]
[html head meta]
[html head title]
[html body]
[html body header]
[html body header h1]
[html body header h1 a]
[html body main]
[html body main h2]
[html body main p]
[html body footer]
[html body footer a]
`

func TestOutline(t *testing.T) {
	var tests = []struct {
		path string
		out  string
	}{
		{"./testdata/short.html", short},
		{"./testdata/long.html", long},
	}

	for _, tt := range tests {
		data, err := os.ReadFile(tt.path)
		if err != nil {
			log.Fatal(err)
		}
		doc, err := html.Parse(strings.NewReader(string(data)))
		if err != nil {
			log.Fatal(err)
		}
		out := new(bytes.Buffer)
		testname := fmt.Sprintf("%s", tt.path)
		t.Run(testname, func(t *testing.T) {
			outline(out, nil, doc)
			if out.String() != tt.out {
				t.Errorf("test for the function outline has failed - results not match\nGot:\n%v\nExpected:\n%v", out.String(), tt.out)
			}
		})
	}
}
