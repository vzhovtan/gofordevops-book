package unmarshalling

import (
	"encoding/json"
)

type InfoInterfaces struct {
	Interfaces map[string]Interface
}
type Interface struct {
	Description string
	Speed       string
	MTU         string
	OperStatus  string
	AdminStatus string
}
type deviceInterface struct {
	TABLEInterface struct {
		ROWInterface struct {
			Interface              string `json:"interface"`
			State                  string `json:"state"`
			AdminState             string `json:"admin_state"`
			ShareState             string `json:"share_state"`
			EthBundle              string `json:"eth_bundle"`
			EthHwDesc              string `json:"eth_hw_desc"`
			EthHwAddr              string `json:"eth_hw_addr"`
			EthBiaAddr             string `json:"eth_bia_addr"`
			Desc                   string `json:"desc"`
			EthMtu                 string `json:"eth_mtu"`
			EthBw                  string `json:"eth_bw"`
			EthBwStr               string `json:"eth_bw_str"`
			EthDly                 string `json:"eth_dly"`
			EthReliability         string `json:"eth_reliability"`
			EthTxload              string `json:"eth_txload"`
			EthRxload              string `json:"eth_rxload"`
			Encapsulation          string `json:"encapsulation"`
			Medium                 string `json:"medium"`
			EthMode                string `json:"eth_mode"`
			EthDuplex              string `json:"eth_duplex"`
			EthSpeed               string `json:"eth_speed"`
			EthMedia               string `json:"eth_media"`
			EthBeacon              string `json:"eth_beacon"`
			EthAutoneg             string `json:"eth_autoneg"`
			EthInFlowctrl          string `json:"eth_in_flowctrl"`
			EthOutFlowctrl         string `json:"eth_out_flowctrl"`
			EthMdix                string `json:"eth_mdix"`
			EthRatemode            string `json:"eth_ratemode"`
			EthSwtMonitor          string `json:"eth_swt_monitor"`
			EthEthertype           string `json:"eth_ethertype"`
			EthEeeState            string `json:"eth_eee_state"`
			EthAdminFecState       string `json:"eth_admin_fec_state"`
			EthOperFecState        string `json:"eth_oper_fec_state"`
			EthLinkFlapped         string `json:"eth_link_flapped"`
			EthClearCounters       string `json:"eth_clear_counters"`
			EthResetCntr           string `json:"eth_reset_cntr"`
			EthLoadInterval1Rx     string `json:"eth_load_interval1_rx"`
			EthInrate1Bits         string `json:"eth_inrate1_bits"`
			EthInrate1Pkts         string `json:"eth_inrate1_pkts"`
			EthLoadInterval1Tx     string `json:"eth_load_interval1_tx"`
			EthOutrate1Bits        string `json:"eth_outrate1_bits"`
			EthOutrate1Pkts        string `json:"eth_outrate1_pkts"`
			EthInrate1SummaryBits  string `json:"eth_inrate1_summary_bits"`
			EthInrate1SummaryPkts  string `json:"eth_inrate1_summary_pkts"`
			EthOutrate1SummaryBits string `json:"eth_outrate1_summary_bits"`
			EthOutrate1SummaryPkts string `json:"eth_outrate1_summary_pkts"`
			EthLoadInterval2Rx     string `json:"eth_load_interval2_rx"`
			EthInrate2Bits         string `json:"eth_inrate2_bits"`
			EthInrate2Pkts         string `json:"eth_inrate2_pkts"`
			EthLoadInterval2Tx     string `json:"eth_load_interval2_tx"`
			EthOutrate2Bits        string `json:"eth_outrate2_bits"`
			EthOutrate2Pkts        string `json:"eth_outrate2_pkts"`
			EthInrate2SummaryBits  string `json:"eth_inrate2_summary_bits"`
			EthInrate2SummaryPkts  string `json:"eth_inrate2_summary_pkts"`
			EthOutrate2SummaryBits string `json:"eth_outrate2_summary_bits"`
			EthOutrate2SummaryPkts string `json:"eth_outrate2_summary_pkts"`
			EthInucast             string `json:"eth_inucast"`
			EthInmcast             string `json:"eth_inmcast"`
			EthInbcast             string `json:"eth_inbcast"`
			EthInpkts              string `json:"eth_inpkts"`
			EthInbytes             string `json:"eth_inbytes"`
			EthJumboInpkts         string `json:"eth_jumbo_inpkts"`
			EthStormSupp           string `json:"eth_storm_supp"`
			EthRunts               string `json:"eth_runts"`
			EthGiants              string `json:"eth_giants"`
			EthCrc                 string `json:"eth_crc"`
			EthNobuf               string `json:"eth_nobuf"`
			EthInerr               string `json:"eth_inerr"`
			EthFrame               string `json:"eth_frame"`
			EthOverrun             string `json:"eth_overrun"`
			EthUnderrun            string `json:"eth_underrun"`
			EthIgnored             string `json:"eth_ignored"`
			EthWatchdog            string `json:"eth_watchdog"`
			EthBadEth              string `json:"eth_bad_eth"`
			EthBadProto            string `json:"eth_bad_proto"`
			EthInIfdownDrops       string `json:"eth_in_ifdown_drops"`
			EthDribble             string `json:"eth_dribble"`
			EthIndiscard           string `json:"eth_indiscard"`
			EthInpause             string `json:"eth_inpause"`
			EthStompedCrc          string `json:"eth_stomped_crc"`
			EthOutucast            string `json:"eth_outucast"`
			EthOutmcast            string `json:"eth_outmcast"`
			EthOutbcast            string `json:"eth_outbcast"`
			EthOutpkts             string `json:"eth_outpkts"`
			EthOutbytes            string `json:"eth_outbytes"`
			EthJumboOutpkts        string `json:"eth_jumbo_outpkts"`
			EthOuterr              string `json:"eth_outerr"`
			EthColl                string `json:"eth_coll"`
			EthDeferred            string `json:"eth_deferred"`
			EthLatecoll            string `json:"eth_latecoll"`
			EthLostcarrier         string `json:"eth_lostcarrier"`
			EthNocarrier           string `json:"eth_nocarrier"`
			EthBabbles             string `json:"eth_babbles"`
			EthOutdiscard          string `json:"eth_outdiscard"`
			EthOutpause            string `json:"eth_outpause"`
		} `json:"ROW_interface"`
	} `json:"TABLE_interface"`
}

func ParseInterfaces(js []byte) (*InfoInterfaces, error) {
	var devInt deviceInterface
	intfs := map[string]Interface{}
	err := json.Unmarshal(js, &devInt)
	if err != nil {
		return nil, err
	}
	intfs[devInt.TABLEInterface.ROWInterface.Interface] = Interface{
		Description: devInt.TABLEInterface.ROWInterface.Desc,
		Speed:       devInt.TABLEInterface.ROWInterface.EthBw,
		MTU:         devInt.TABLEInterface.ROWInterface.EthMtu,
		OperStatus:  devInt.TABLEInterface.ROWInterface.State,
		AdminStatus: devInt.TABLEInterface.ROWInterface.AdminState,
	}
	return &InfoInterfaces{Interfaces: intfs}, nil
}
