package extraction

import (
	"encoding/json"
)

type DeviceInterface struct {
	TABLEInterface struct {
		ROWInterface struct {
			Interface  string `json:"interface"`
			State      string `json:"state"`
			AdminState string `json:"admin_state"`
			Desc       string `json:"desc"`
			EthMtu     string `json:"eth_mtu"`
			EthBw      string `json:"eth_bw"`
		} `json:"ROW_interface"`
	} `json:"TABLE_interface"`
}

type InfoInterface struct {
	Description string
	Speed       string
	MTU         string
	OperStatus  string
	AdminStatus string
}

type InfoInterfaces struct {
	Interfaces map[string]InfoInterface
}

func ParseDeviceInterface(js []byte) (*InfoInterfaces, error) {
	var devint DeviceInterface
	intfs := map[string]InfoInterface{}
	err := json.Unmarshal(js, &devint)
	if err != nil {
		return nil, err
	}
	intfs[devint.TABLEInterface.ROWInterface.Interface] = InfoInterface{
		Description: devint.TABLEInterface.ROWInterface.Desc,
		Speed:       devint.TABLEInterface.ROWInterface.EthBw,
		MTU:         devint.TABLEInterface.ROWInterface.EthMtu,
		OperStatus:  devint.TABLEInterface.ROWInterface.State,
		AdminStatus: devint.TABLEInterface.ROWInterface.AdminState,
	}
	return &InfoInterfaces{Interfaces: intfs}, nil
}
