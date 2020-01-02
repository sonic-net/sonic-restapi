package main

import (
    "fmt"
    "arpthrift"
    "git.apache.org/thrift.git/lib/go/thrift"
)

func main() {
    transport, err := thrift.NewTServerSocket("localhost:9091")

    if err != nil {
        fmt.Println("Error opening socket:", err)
        return
    }

    defer transport.Close()
    if err := transport.Open(); err != nil {
        fmt.Println("Error opening transport", err)
        return
    }

    handler := NewArpHandler()
    processor := arp.NewArpResponderProcessor(handler)
    server := thrift.NewTSimpleServer2(processor, transport)

    fmt.Println("Starting server...")
    server.Serve()
}

type ArpHandler struct {
}

func NewArpHandler() *ArpHandler {
    return &ArpHandler{}
}

func (p *ArpHandler) AddInterface(iface_name string) (r bool, err error) {
    fmt.Printf("add_interface(%v)\n", iface_name)
    r = true
    return
}

func (p *ArpHandler) DelInterface(iface_name string) (r bool, err error) {
    fmt.Printf("del_interface(%v)\n", iface_name)
    r = true
    return
}

func (p *ArpHandler) AddIP(iface_name string, stag arp.VlanTagT, ctag arp.VlanTagT, ip arp.Ip4T) (r bool, err error) {
    fmt.Printf("add_ip(%v, %+v, %v, %v)\n", iface_name, stag, ctag, ip)
    r = true
    return
}

func (p *ArpHandler) DelIP(iface_name string, stag arp.VlanTagT, ctag arp.VlanTagT) (r bool, err error) {
    fmt.Printf("del_ip(%v, %+v, %+v)\n", iface_name, stag, ctag)
    r = true
    return
}

func (p *ArpHandler) RequestMac(requests []*arp.ReqTuplesT) (r []*arp.RepTupleT, err error) {
    fmt.Printf("request_mac(%v)\n", requests)

    for _, req := range requests {
        if req.IP == 203569230 {
            r = append(r, &arp.RepTupleT{
                Request: req.Tuples[int(req.Index) % len(req.Tuples)],
                Index: req.Index,
                Mac: arp.MacT([]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12}),
                IsFound: true,
            })
        } else if req.IP == 1510744632 {
            r = append(r, &arp.RepTupleT{
                Request: req.Tuples[int(req.Index) % len(req.Tuples)],
                Index: req.Index,
                Mac: arp.MacT([]byte{0x34, 0x56, 0x78, 0x90, 0x12, 0x34}),
                IsFound: true,
            })
        }
    }

    return
}
