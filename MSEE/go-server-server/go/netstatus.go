package mseeserver

import (
    "io/ioutil"
    "os"
    "strings"
)

func GetAllNetworkStatuses() (operstatuses map[string]bool, err error) {
    operstatuses = make(map[string]bool)

    fi, err := ioutil.ReadDir("/sys/class/net")
    if err != nil {
        return
    }

    for _, f := range fi {
        var ns bool
        ns, err = GetNetworkStatus(f.Name())
        if err != nil {
            return
        }
        operstatuses[f.Name()] = ns
    }

    return
}

func GetNetworkStatus(port string) (operstatus bool, err error) {
    b, err := ioutil.ReadFile("/sys/class/net/" + port + "/operstate")
    if err != nil {
        return
    }

    if string(b) == "up\n" {
        operstatus = true
    } else {
        operstatus = false
    }

    return
}

func GetPorts(prefix string) (ports []string, err error) {
    fi, err := ioutil.ReadDir("/sys/class/net")
    if err != nil {
        return
    }

    for _, f := range fi {
        if strings.HasPrefix(f.Name(), prefix) {
            ports = append(ports, f.Name())
        }
    }

    return
}

func PortExists(port string) bool {
    _, err := os.Stat("/sys/class/net/" + port)
    return err == nil
}
