package main

import (
    "fmt"
    "mseethrift"
    "git.apache.org/thrift.git/lib/go/thrift"
)

func main() {
    transport, err := thrift.NewTServerSocket("localhost:9090")

    if err != nil {
        fmt.Println("Error opening socket:", err)
        return
    }

    defer transport.Close()
    if err := transport.Open(); err != nil {
        fmt.Println("Error opening transport", err)
        return
    }

    handler := NewMSEEHandler()
    processor := msee.NewMSEEProcessor(handler)
    server := thrift.NewTSimpleServer2(processor, transport)

    fmt.Println("Starting server...")
    server.Serve()
}

type MSEEHandler struct {
}

func NewMSEEHandler() *MSEEHandler {
    return &MSEEHandler{}
}

func (p *MSEEHandler) InitDpdkPort(nb_customer_ports msee.MseePortCountT, mac_addr msee.MseeMacT, ipv4_loaddr msee.MseeIp4T, ipv6_loaddr *msee.MseeIp6T) (r msee.ResultT, err error) {
    fmt.Printf("init_dpdk_port(%v, %v, %+v, %+v)\n", nb_customer_ports, mac_addr, ipv4_loaddr, ipv6_loaddr)
    r = msee.ResultT_OK
    return
}

func (p *MSEEHandler) AddPortToVrf(vrf_id msee.MseeVrfIDT, port msee.MseePortT, outer_vlan msee.MseeVlanT, inner_vlan msee.MseeVlanT) (r msee.ResultT, err error) {
    fmt.Printf("add_port_to_vrf(%v, %v, %v, %v)\n", vrf_id, port, outer_vlan, inner_vlan)
    r = msee.ResultT_ADDED
    return
}

func (p *MSEEHandler) DeletePortFromVrf(port msee.MseePortT, outer_vlan msee.MseeVlanT, inner_vlan msee.MseeVlanT) (r msee.ResultT, err error) {
    fmt.Printf("delete_port_from_vrf(%v, %v, %v)\n", port, outer_vlan, inner_vlan)
    r = msee.ResultT_REMOVED
    return
}

func (p *MSEEHandler)MapVniToVrf(vni msee.MseeVniT, vrf_id msee.MseeVrfIDT) (r msee.ResultT, err error) {
    fmt.Printf("map_vni_to_vrf(%v, %v)\n", vni, vrf_id)
    r = msee.ResultT_ADDED
    return
}

func (p *MSEEHandler) UnmapVniToVrf(vni msee.MseeVniT) (r msee.ResultT, err error) {
    fmt.Printf("unmap_vni_to_vrf(%v)\n", vni)
    r = msee.ResultT_REMOVED
    return
}

func (p *MSEEHandler) AddEncapRoute(vrf_id msee.MseeVrfIDT, dst_vm_ip_prefix *msee.MseeIPPrefixT, dst_host_ip *msee.MseeIPAddressT, dst_mac_address msee.MseeMacT, vni msee.MseeVniT, port msee.MseeUDPPortT) (r msee.ResultT, err error) {
    fmt.Printf("add_encap_route(%v, %v, %v, %v, %v, %v)\n", vrf_id, dst_vm_ip_prefix, dst_host_ip, dst_mac_address, vni, uint16(port))
    r = msee.ResultT_ADDED
    return
}

func (p *MSEEHandler) DeleteEncapRoute(vrf_id msee.MseeVrfIDT, dst_vm_ip_prefix *msee.MseeIPPrefixT) (r msee.ResultT, err error) {
    fmt.Printf("delete_encap_route(%v, %v)\n", vrf_id, dst_vm_ip_prefix)
    r = msee.ResultT_REMOVED
    return
}

func (p *MSEEHandler) AddDecapRoute(vrf_id msee.MseeVrfIDT, dst_ip_prefix *msee.MseeIPPrefixT, mac msee.MseeMacT, port msee.MseePortT, outer_vlan msee.MseeVlanT, inner_vlan msee.MseeVlanT) (r msee.ResultT, err error) {
    fmt.Printf("add_decap_route(%v, %v, %v, %v, %v, %v)\n", vrf_id, dst_ip_prefix, mac, port, outer_vlan, inner_vlan)
    r = msee.ResultT_ADDED
    return
}

func (p *MSEEHandler) DeleteDecapRoute(vrf_id msee.MseeVrfIDT, dst_ip_prefix *msee.MseeIPPrefixT) (r msee.ResultT, err error) {
    fmt.Printf("delete_decap_route(%v, %v)\n", vrf_id, dst_ip_prefix)
    r = msee.ResultT_REMOVED
    return
}

func (p *MSEEHandler) GetCounters(group msee.MseeGroupT) (r msee.CountersT, err error) {
    fmt.Printf("get_counters(%v)\n", group)
    if len(group) > 0 {
        r = msee.CountersT{group: map[msee.MseeCounterName]int64{"0.decap_ok": 1, "0.encap_lpm_not_found": 2}}
    } else {
        r = msee.CountersT{"dpdk.switch_ports": map[msee.MseeCounterName]int64{"0.decap_ok": 1, "0.encap_lpm_not_found": 2, "100.encap_lpm_not_found": 2, "foo": 2}, "dpdk.total": map[msee.MseeCounterName]int64{"0.foo": 3, "bar": 4}}
    }
    return
}

func (p *MSEEHandler) GetStatistics(group msee.MseeGroupT) (r msee.StatisticsT, err error) {
    fmt.Printf("get_statistics(%v)\n", group)
    if len(group) > 0 {
        r = msee.StatisticsT{group: map[msee.MseeStatisticsName]int64{"foo": 1, "bar": 2}}
    } else {
        r = msee.StatisticsT{"rings": map[msee.MseeStatisticsName]int64{"foo": 1, "bar": 2}, "mempools": map[msee.MseeStatisticsName]int64{"foo": 3, "bar": 4}}
    }
    return
}

func (p *MSEEHandler) GetHist() (r msee.HistT, err error) {
    fmt.Printf("get_hist()\n")
    r = msee.HistT{0: map[int8]float64{1: 1.1, 2: 2.1}, 1: map[int8]float64{1: 1.1, 2: 2.1}}
    return
}