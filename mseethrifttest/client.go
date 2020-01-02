package main

import (
    "fmt"
    "mseethrift"
    "git.apache.org/thrift.git/lib/go/thrift"
)

func main() {
    transport, err := thrift.NewTSocket("localhost:9090")

    if err != nil {
        fmt.Println("Error opening socket:", err)
        return
    }

    defer transport.Close()
    if err := transport.Open(); err != nil {
        fmt.Println("Error opening transport", err)
        return
    }

    protocol := thrift.NewTBinaryProtocolTransport(transport)

    client := msee.NewMSEEClientProtocol(transport, protocol, protocol)

    result, err := client.SetSwitchAddr(0x4c7625f52a80, 0x194e9865)
    checkResult("SetSwitchAddr", result, err)

    result, err = client.AddPortToVrf("Ethernet0", "1234")
    checkResult("AddPortToVrf", result, err)

    result, err = client.AddVroute("1234", 0xc0a80204, 32, 17291, 0x0a508818)
    checkResult("AddVroute", result, err)

    result, err = client.AddVroute("1234", 0xc0a80206, 32, 17291, 0x0a508c17)
    checkResult("AddVroute", result, err)

    result, err = client.AddVroute("1234", 0xc0a80205, 32, 17291, 0x0a508819)
    checkResult("AddVroute", result, err)
}

func checkResult(method string, result bool, err error) {
    if err != nil {
        fmt.Printf("Error %s\n", method)
    } else {
        fmt.Printf("Result of %s: %v\n", method, result)
    }
}
