package mseeserver

import "flag"

var ResetFlag = flag.Bool("reset", false, "Resets cache DB")
var SWMacAddrFlag = flag.String("swmacaddr", "", "Switch Mac address")
var DPDKMacAddrFlag = flag.String("dpdkmacaddr", "", "DPDK Mac address")
var LoAddr4Flag = flag.String("loaddr4", "", "IPv4 loopback addresses")
var LoAddr6Flag = flag.String("loaddr6", "", "IPv6 loopback addresses")
var ThriftHostFlag = flag.String("dpdkthrift", "localhost:9090", "hostname:port for DPDK Thrift server")
var ARPHostFlag = flag.String("arpthrift", "localhost:9091", "hostname:port for ARP Thrift server")
var LogLevelFlag = flag.String("loglevel", "info", "Set's minimum log level, valid values are: trace, debug, info, warning, error, alert")
var LogFileFlag = flag.String("logfile", "/dev/stderr", "Set's the output for the log")
var VlanStartFlag = flag.Int("vlanstart", 3840, "Starting number for port VLANs")
var DpdkPortFlag = flag.String("dpdkport", "Ethernet0", "Port that DPDK server is attached to")
var PortChannelPortsFlag = flag.String("pcports", "", "Comma separated list of ports used for port channels")
var PortsFlag = flag.String("ports", "", "Comma separated list of ports to isolate")
var HttpFlag = flag.Bool("enablehttp", true, "Enable http endpoint")
var HttpsFlag = flag.Bool("enablehttps", false, "Enable https endpoint")
var ClientCertFlag = flag.String("clientcert", "", "Client cert file")
var ClientCertCommonNameFlag = flag.String("clientcertcommonname", "SonicCLient", "Comma separated list of trusted common names in the client cert file")
var ServerCertFlag = flag.String("servercert", "", "Server cert file")
var ServerKeyFlag = flag.String("serverkey", "", "Server key file")

