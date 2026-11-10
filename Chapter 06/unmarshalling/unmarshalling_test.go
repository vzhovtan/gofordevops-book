package unmarshalling_test

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"testing"
	"unmarshalling"
)

func TestParseInterfaces(t *testing.T) {
	tests := []struct {
		name string
		want *unmarshalling.InfoInterfaces
	}{
		{
			name: "Parsing device interfaces output",
			want: &unmarshalling.InfoInterfaces{
				Interfaces: map[string]unmarshalling.Interface{
					"Ethernet1/1": {
						Description: "another.device:peering-interface",
						Speed:       "100000000",
						MTU:         "9216",
						OperStatus:  "up",
						AdminStatus: "up",
					},
				},
			},
		},
	}
	data, err := os.ReadFile(filepath.Join("testdata", "interface_output"))
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unmarshalling.ParseInterfaces(data)
			if err != nil {
				t.Errorf("ParseInterface got error: %v, want nil", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseInterface returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
