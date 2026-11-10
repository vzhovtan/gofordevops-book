package extraction

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"testing"
)

func TestParseInterfaces(t *testing.T) {
	tests := []struct {
		name string
		want *InfoInterfaces
	}{
		{
			name: "Parsing device interfaces output",
			want: &InfoInterfaces{
				Interfaces: map[string]InfoInterface{
					"Ethernet1/3": {
						Description: "CustomLab",
						Speed:       "1000000",
						MTU:         "1500",
						OperStatus:  "up",
						AdminStatus: "up",
					},
				},
			},
		},
	}
	data, err := os.ReadFile(filepath.Join("testdata", "interface.json"))
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDeviceInterface(data)
			if err != nil {
				t.Errorf("ParseDeviceInterface got error: %v, want nil", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseDeviceInterface returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
