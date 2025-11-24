package appio_test

import (
	"appio"
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var TestBase = appio.Base{
	Device_id:        "TEST_DEVICE_ID",
	Management_ipv6:  "2001:22:22:22",
	Management_ip:    "10.10.10.10",
	Network_password: "PW",
	Tacacs_key:       "TACACS_KEY",
	Bootstrap:        false,
	Snmp_community:   "TEST_COMMUNITY",
	Mgmt_lo1_id:      "1.1.1.1",
	Mgmt_lo2_id:      "2.2.2.2",
	Mgmt_lo3_id:      "3.3.3.3",
	Mgmt_iso_id:      "ISO-ID_TEST",
	Asn_id:           "6530_TEST",
	Rd_id:            "123:456",
}

var TestL2 = appio.L2info{
	Int_type:         "uplink",
	Int_id:           "111",
	Int_desc:         "TEST_DESCRIPTION",
	Rfc_address:      "192.168.1.1",
	Ipv6_rfc_address: "2001:11:11",
}

var TestL3 = appio.L3info{
	Neighbor_id:   "TEST_NE_ID",
	Neighbor_desc: "TEST_NE_DESCR",
}

var TestDeviceResult = appio.Device{
	Base: TestBase,
	Ports: []appio.L2info{
		TestL2,
	},
	Interfaces: []appio.L3info{
		TestL3,
	},
}

func TestBuildDevice(t *testing.T) {
	dev := appio.BuildDevice(TestBase, TestL2, TestL3)
	if !cmp.Equal(*dev, TestDeviceResult) {
		t.Errorf("test BuildDevice Failed - error")
	}
}

func TestBuildConfig(t *testing.T) {
	file, err := os.Open("./testdata/testresult")
	if err != nil {
		t.Fatalf("Error opening test file - %v\n", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		t.Fatalf("Error getting test file size - %v\n", err)
	}
	fileSize := fileInfo.Size()

	want := make([]byte, fileSize)
	_, err = file.Read(want)
	if err != nil {
		t.Fatalf("Error reading a test file - %v\n", err)
	}

	var b bytes.Buffer

	dev := appio.BuildDevice(TestBase, TestL2, TestL3)
	err = appio.BuildConfig(&b, dev)

	if err != nil {
		t.Fatalf("Error Building Config - %v\n", err)
	}

	data := b.Bytes()
	re := regexp.MustCompile(`\n|\s`)
	expected := re.ReplaceAll(want, []byte(""))
	actual := re.ReplaceAll(data, []byte(""))
	if !cmp.Equal(expected, actual) {
		t.Errorf("test BuildConfig Failed - results not match\nGot:\n%v\nExpected:\n%v\n", string(actual), string(expected))
	}
}
