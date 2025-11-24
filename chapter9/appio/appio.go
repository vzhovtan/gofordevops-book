package appio

import (
	"io"
	"text/template"
)

type Base struct {
	Device_id        string
	Management_ipv6  string
	Management_ip    string
	Network_password string
	Tacacs_key       string
	Bootstrap        bool
	Snmp_community   string
	Mgmt_lo1_id      string
	Mgmt_lo2_id      string
	Mgmt_lo3_id      string
	Mgmt_iso_id      string
	Asn_id           string
	Rd_id            string
}

type L2info struct {
	Int_type         string
	Int_id           string
	Int_desc         string
	Rfc_address      string
	Ipv6_rfc_address string
}
type L3info struct {
	Neighbor_id   string
	Neighbor_desc string
}

type Device struct {
	Base       Base
	Ports      []L2info
	Interfaces []L3info
}

func BuildDevice(base Base, port L2info, intfs L3info) *Device {
	return &Device{
		Base:       base,
		Ports:      []L2info{port},
		Interfaces: []L3info{intfs},
	}
}

func BuildConfig(out io.Writer, dev *Device) error {
	t, err := template.ParseFiles("conf.tmpl")
	if err != nil {
		return err
	}
	err = t.Execute(out, dev)
	if err != nil {
		return err
	}
	return nil
}
