package extraction

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"testing"
)

func TestParseISIS(t *testing.T) {
	tests := []struct {
		name string
		want *InfoISIS
	}{
		{
			name: "Parsing device API output",
			want: &InfoISIS{
				Instance: "internal",
				Neighbors: []Neighbor{
					Neighbor{
						SystemID:  "sw1",
						Interface: "Ethernet1/49",
					},
					Neighbor{
						SystemID:  "sw2",
						Interface: "Ethernet1/51",
					},
				},
			},
		},
	}
	data, err := os.ReadFile(filepath.Join("testdata", "isis.json"))
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseISIS(data)
			if err != nil {
				t.Errorf("ParseISIS got error: %v, want nil", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseISIS returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
