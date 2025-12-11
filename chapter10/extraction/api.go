package extraction

import (
	"encoding/json"
)

type Neighbor struct {
	SystemID  string
	SNPA      string
	AdjNumber int64
	AdjState  string
	IntAuth   string
	Interface string
	Area      string
	VRF       string
	Instance  string
}

type InfoISIS struct {
	Neighbors           []Neighbor
	NeighborsByID       map[string]Neighbor
	SystemID            string
	PassiveInterface    string
	GenerateWideMetrics string
	AcceptWideMetrics   string
	SupportedRoutes     string
	AreaIds             []string
	Instance            string
}
type deviceISIS struct {
	TABLEProcessTag struct {
		ROWProcessTag struct {
			ProcessTagOut string `json:"process-tag-out"`
			TABLEVrf      struct {
				ROWVrf struct {
					VrfNameOut      string `json:"vrf-name-out"`
					AdjSummaryOut   string `json:"adj-summary-out"`
					AdjInterfaceOut string `json:"adj-interface-out"`
					TABLEProcessAdj struct {
						ROWProcessAdj []struct {
							AdjSysNameOut   string `json:"adj-sys-name-out"`
							AdjSysIDOut     string `json:"adj-sys-id-out"`
							AdjUsageOut     string `json:"adj-usage-out"`
							AdjStateOut     string `json:"adj-state-out"`
							AdjHoldTimeOut  string `json:"adj-hold-time-out"`
							AdjIntfNameOut  string `json:"adj-intf-name-out"`
							AdjDetailSetOut string `json:"adj-detail-set-out"`
						} `json:"ROW_process_adj"`
					} `json:"TABLE_process_adj"`
				} `json:"ROW_vrf"`
			} `json:"TABLE_vrf"`
		} `json:"ROW_process_tag"`
	} `json:"TABLE_process_tag"`
}

func ParseISIS(js []byte) (*InfoISIS, error) {
	var isis deviceISIS
	neighbors := []Neighbor{}
	err := json.Unmarshal(js, &isis)
	if err != nil {
		return nil, err
	}
	instance := isis.TABLEProcessTag.ROWProcessTag.ProcessTagOut
	for _, peer := range isis.TABLEProcessTag.ROWProcessTag.TABLEVrf.ROWVrf.TABLEProcessAdj.ROWProcessAdj {
		nei := peer.AdjSysNameOut
		inter := peer.AdjIntfNameOut
		neighbors = append(neighbors, Neighbor{
			SystemID:  nei,
			Interface: inter})
	}
	return &InfoISIS{
		Neighbors: neighbors,
		Instance:  instance,
	}, nil
}
