package mseeserver

import (
    "arpthrift"
    "fmt"
    "github.com/go-redis/redis"
    "log"
    "mseethrift"
    "net"
    "net/http"
    "strconv"
    "strings"
    "swsscommon"
    "regexp"
    "errors"

    "github.com/gorilla/mux"
)

func StateHeartbeatGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    output := HeartbeatReturnModel{
        ServerVersion: ServerAPIVersion,
        ResetGUID: ServerResetGuid,
        ResetTime: ServerResetTime,
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    WriteRequestResponse(w, *configSnapshot, http.StatusOK)
}

func ConfigInterfacePortPortDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    addr, _, err := CacheGetPortAddr(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"port"}, "")
        return
    }

    table := vars["port"]+":"+addr
    kv, err := SwssGetKVs("IGNORE_INTF_TABLE:"+table)
    if err != nil || !PortExists(vars["port"]) {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"port"}, "")
        return
    }

    vrfStr := kv["vrf_id"]
    vrfID, err := strconv.Atoi(vrfStr)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "vrf_id saved in database is not an integer")
        return
    }

    portID, err := PortToPortID(vars["port"])

    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"port"}, "Port not allowed")
        return
    }

    // If we didn't return an error at PortToPortID we won't now
    vlanID, _ := PortToVlanID(vars["port"])

    mseeVrfID := msee.MseeVrfIDT(vrfID)
    mseePort := msee.MseePortT(portID)

    mseePrefix := GetThriftIPPrefix(addr)
    
    ret, err := mseeClient.DeleteDecapRoute(mseeVrfID, &mseePrefix)
    str := fmt.Sprintf("trace: thrift: delete_decap_route(%v, %v) = (%v, %v)", mseeVrfID, mseePrefix, ret, err)

    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
        return
    }

    ret, err = mseeClient.DeletePortFromVrf(mseePort, 0, 0)
    str = fmt.Sprintf("trace: thrift: delete_port_from_vrf(%v, 0, 0) = (%v, %v)", mseePort, ret, err)

    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
        return
    }

    // ret, err = arpClient.DelInterface(vars["port"])
    // log.Printf("trace: thrift: del_interface(%v) = (%v, %v)", vars["port"], ret, err)

    // if err != nil || !ret {
    //  WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
    //  return
    // }

    pt := swsscommon.NewProducerStateTable(swssDB, "IGNORE_INTF_TABLE")
    defer pt.Delete()

    pt.Del(table, "DEL", "")

    err = CacheDeletePortAddr(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    spt := swsscommon.NewProducerStateTable(swssDB, "SPOOF_GUARD_TABLE")
    defer spt.Delete()

    spt.Del(vars["port"], "DEL", "")

    vpt := swsscommon.NewProducerStateTable(swssDB, "VLAN_MEMBER_TABLE")
    defer vpt.Delete()

    vpt.Del(fmt.Sprintf("Vlan%d:%s", vlanID, *DpdkPortFlag), "DEL", "")

    if _, ok := configSnapshot.VrfMap[vrfID]; ok {
        delete(configSnapshot.VrfMap[vrfID].PortMap, vars["port"])
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfacePortPortGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    addr, macaddr, err := CacheGetPortAddr(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"port"}, "")
        return
    }

    kv, err := SwssGetKVs("IGNORE_INTF_TABLE:" + vars["port"] + ":" + addr)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }

    vrfID, err := strconv.Atoi(kv["vrf_id"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    output := PortReturnModel{
        Port: vars["port"],
        Attr: PortModel{
            Addr:  addr,
            VrfID: vrfID,
            MACAddress: macaddr,
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfacePortPortPut(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    var attr PortModel

    err := ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    _, err = CacheGetVrfName(attr.VrfID)
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vrf_id"}, "")
        return
    }

    _, _, err = ParseIPPrefix(attr.Addr)

    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"addr"}, "Not a valid IP prefix")
        return
    }

    if !PortExists(vars["port"]) {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"port"}, "")
        return
    }

    portID, err := PortToPortID(vars["port"])

    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"port"}, "Port not allowed")
        return
    }

    vlanID := portID + *VlanStartFlag

    pt := swsscommon.NewProducerStateTable(swssDB, "IGNORE_INTF_TABLE")
    defer pt.Delete()

    spt := swsscommon.NewProducerStateTable(swssDB, "SPOOF_GUARD_TABLE")
    defer spt.Delete()

    vpt := swsscommon.NewProducerStateTable(swssDB, "VLAN_MEMBER_TABLE")
    defer vpt.Delete()

    oldAddr, oldmacaddr, err := CacheGetPortAddr(vars["port"])

    if err != redis.Nil {
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        table := vars["port"]+":"+oldAddr
        kv, err := SwssGetKVs("IGNORE_INTF_TABLE:" + table)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        if kv == nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "The key IGNORE_INTF_TABLE:" + table + " could not be found")
            return
        }

        vrfStr := kv["vrf_id"]
        vrfID, err := strconv.Atoi(vrfStr)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        if vrfID != attr.VrfID {
            WriteRequestError(w, http.StatusMethodNotAllowed, "Method not allowed", []string{}, "vrf_id cannot be updated")
            return
        }

        macOld, _ := net.ParseMAC(oldmacaddr)
        macNew, _ := net.ParseMAC(attr.MACAddress)

        if  MacToInt64(macOld) != MacToInt64(macNew) {
            WriteRequestError(w, http.StatusMethodNotAllowed, "Method not allowed", []string{}, "mac_address cannot be updated")
            return
        }

        ipOld, netOld, _ := net.ParseCIDR(oldAddr)
        ipNew, netNew, _ := net.ParseCIDR(attr.Addr)
        lenOld, _ := netOld.Mask.Size()
        lenNew, _ := netNew.Mask.Size()

        if !ipOld.Equal(ipNew) || !netOld.IP.Equal(netNew.IP) || lenOld != lenNew {
            WriteRequestError(w, http.StatusMethodNotAllowed, "Method not allowed", []string{}, "addr cannot be updated")
            return
        }

        spt.Del(vars["port"], "DEL", "")
    } else {
        mseeVrfID := msee.MseeVrfIDT(attr.VrfID)
        mseePort := msee.MseePortT(portID)

        ret, err := mseeClient.AddPortToVrf(mseeVrfID, mseePort, 0, 0)
        str := fmt.Sprintf("trace: thrift: add_port_to_vrf(%v, %v, 0, 0) = (%v, %v)", mseeVrfID, mseePort, ret, err)

        if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
            return
        }

        err = CacheSetPortAddr(vars["port"], attr.Addr, attr.MACAddress)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        pt.Set(vars["port"]+":"+attr.Addr, map[string]string{
            "vrf_id": strconv.Itoa(attr.VrfID),
        }, "SET", "")

        kv, err := SwssGetKVs(fmt.Sprintf("VLAN_MEMBER_TABLE:Vlan%d:%s", vlanID, *DpdkPortFlag))
        if kv == nil {
            vpt.Set(fmt.Sprintf("Vlan%d:%s", vlanID, *DpdkPortFlag), map[string]string{
                "tagging_mode": "tagged",
            }, "SET", "")
        }

        mseePrefix := GetThriftIPPrefix(attr.Addr)
        mac, _ := net.ParseMAC(attr.MACAddress)
        mseeMACAddress := msee.MseeMacT(MacToInt64(mac))

        ret, err = mseeClient.AddDecapRoute(mseeVrfID, &mseePrefix, mseeMACAddress, mseePort, 0, 0)
        str = fmt.Sprintf("trace: thrift: add_decap_route(%v, %v, %v, %v) = (%v, %v)", mseeVrfID, mseePrefix, mseeMACAddress, mseePort, ret, err)

        if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
            return
        }
    }

    if len(attr.SpoofGuard) > 0 {
        spt.Set(vars["port"], map[string]string{
            "addr_list": strings.Join(attr.SpoofGuard, ","),
        }, "SET", "")
    }

    if _, ok := configSnapshot.VrfMap[attr.VrfID]; ok {
        configSnapshot.VrfMap[attr.VrfID].PortMap[vars["port"]] = attr
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceQinqPortDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    portID, err := PortToPortID(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"port"}, "")
        return
    }

    mseePortID := msee.MseePortT(portID)

    kv, err := SwssGetKVsMulti("QINQ_TABLE:"+vars["port"]+":*")
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    ret, err := arpClient.DelInterface(vars["port"])
    str := fmt.Sprintf("trace: thrift: del_interface(%v) = (%v, %v)", vars["port"], ret, err)
    log.Printf(str)

    if err != nil || !ret {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "QINQ_TABLE")
    defer pt.Delete()

    for k, v := range kv {
        vrfID, err := strconv.Atoi(v["vrf_id"])
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "vrf_id is in wrong format")
            return
        }

        table := k[len("QINQ_TABLE:"):]
        pt.Del(table, "DEL", "")

        tableparts := strings.Split(table, ":")
        stag := tableparts[1]
        ctag := tableparts[2]

        outerVlan, err := strconv.Atoi(stag)
        mseeOuterVlan := msee.MseeVlanT(outerVlan)
        innerVlan, err := strconv.Atoi(ctag)
        mseeInnerVlan := msee.MseeVlanT(innerVlan)

        ret, err := mseeClient.DeletePortFromVrf(mseePortID, mseeOuterVlan, mseeInnerVlan)
        str := fmt.Sprintf("trace: thrift: delete_port_from_vrf(%v, %v, %v) = (%v, %v)", mseePortID, mseeOuterVlan, mseeInnerVlan, ret, err)

        if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
            return
        }

        if _, ok := configSnapshot.VrfMap[vrfID]; ok {
            delete(configSnapshot.VrfMap[vrfID].QinQPortMap, table)
        }
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceQinqPortGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    kv, err := SwssGetKVsMulti("QINQ_TABLE:"+vars["port"]+":*")
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    ports := make([]QInQReturnModel, 0, len(kv))

    for k, kvp := range kv {
        parts := strings.Split(k, ":")

        stag, err := strconv.Atoi(parts[2])
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        ctag, err := strconv.Atoi(parts[3])
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        vrfID, err := strconv.Atoi(kvp["vrf_id"])
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        ports = append(ports, QInQReturnModel{
            Port: vars["port"],
            STag: stag,
            CTag: ctag,
            Attr: QInQModel{
                VrfID:      vrfID,
                PeerIP:     kvp["peer_ip"],
                ProxyArpIP: kvp["proxy_arp_ip"],
                Subnet:     kvp["subnet"],
            },
        })
    }

    WriteRequestResponse(w, ports, http.StatusOK)
}

func ConfigInterfaceQinqPortStagCtagDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    _, _, err := ValidateStagCtag(w, vars["stag"], vars["ctag"])
    if err != nil {
        return
    }

    table := vars["port"]+":"+vars["stag"]+":"+vars["ctag"]
    kv, err := SwssGetKVs("QINQ_TABLE:"+table)
    if err != nil || kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }

    vrfID, err := strconv.Atoi(kv["vrf_id"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "vrf_id is in wrong format")
        return
    }

    portID, err := PortToPortID(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"port"}, "")
        return
    }

    mseePortID := msee.MseePortT(portID)
    outerVlan, err := strconv.Atoi(vars["stag"])
    mseeOuterVlan := msee.MseeVlanT(outerVlan)
    innerVlan, err := strconv.Atoi(vars["ctag"])
    mseeInnerVlan := msee.MseeVlanT(innerVlan)

    ret, err := mseeClient.DeletePortFromVrf(mseePortID, mseeOuterVlan, mseeInnerVlan)
    str := fmt.Sprintf("trace: thrift: delete_port_from_vrf(%v, %v, %v) = (%v, %v)", mseePortID, mseeOuterVlan, mseeInnerVlan, ret, err)

    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
        return
    }

    arpOuterVlan := arp.VlanTagT(outerVlan)
    arpInnerVlan := arp.VlanTagT(innerVlan)

    retarp, err := arpClient.DelIP(vars["port"], arpOuterVlan, arpInnerVlan)
    str = fmt.Sprintf("trace: thrift: del_ip(%v, %v, %v) = (%v, %v)", vars["port"], outerVlan, innerVlan, retarp, err)
    log.Print(str)

    if err != nil || !retarp {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "QINQ_TABLE")
    defer pt.Delete()

    pt.Del(table, "DEL", "")

    if _, ok := configSnapshot.VrfMap[vrfID]; ok {
        delete(configSnapshot.VrfMap[vrfID].QinQPortMap, table)
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceQinqPortStagCtagGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    stag, ctag, err := ValidateStagCtag(w, vars["stag"], vars["ctag"])
    if err != nil {
        return
    }

    kv, err := SwssGetKVs("QINQ_TABLE:"+vars["port"]+":"+vars["stag"]+":"+vars["ctag"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }

    vrfID, err := strconv.Atoi(kv["vrf_id"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    output := QInQReturnModel{
        Port: vars["port"],
        STag: stag,
        CTag: ctag,
        Attr: QInQModel{
            VrfID:      vrfID,
            PeerIP:     kv["peer_ip"],
            ProxyArpIP: kv["proxy_arp_ip"],
            Subnet:     kv["subnet"],
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfaceQinqPortStagCtagPut(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    _, _, err := ValidateStagCtag(w, vars["stag"], vars["ctag"])
    if err != nil {
        return
    }

    var attr QInQModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    portID, err := PortToPortID(vars["port"])
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"port"}, "")
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "QINQ_TABLE")
    defer pt.Delete()

    mseePortID := msee.MseePortT(portID)
    outerVlan, err := strconv.Atoi(vars["stag"])
    mseeOuterVlan := msee.MseeVlanT(outerVlan)
    innerVlan, err := strconv.Atoi(vars["ctag"])
    mseeInnerVlan := msee.MseeVlanT(innerVlan)

    arpOuterVlan := arp.VlanTagT(outerVlan)
    arpInnerVlan := arp.VlanTagT(innerVlan)

    table := vars["port"]+":"+vars["stag"]+":"+vars["ctag"]
    kv, err := SwssGetKVs("QINQ_TABLE:"+table)
    if kv == nil {
        mseeVrfID := msee.MseeVrfIDT(attr.VrfID)

        ret, err := mseeClient.AddPortToVrf(mseeVrfID, mseePortID, mseeOuterVlan, mseeInnerVlan)
        str := fmt.Sprintf("trace: thrift: add_port_to_vrf(%v, %v, %v, %v) = (%v, %v)", mseeVrfID, mseePortID, mseeOuterVlan, mseeInnerVlan, ret, err)

        if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
            return
        }

        retarp, err := arpClient.AddInterface(vars["port"])
        str = fmt.Sprintf("trace: thrift: add_interface(%v) = (%v, %v)", vars["port"], retarp, err)
        log.Print(str)

        if err != nil || !retarp {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
            return
        }
    } else {
        if kv["vrf_id"] != strconv.Itoa(attr.VrfID) {
            WriteRequestError(w, http.StatusMethodNotAllowed, "Method not allowed", []string{}, "vrf_id cannot be updated")
            return
        }

        retarp, err := arpClient.DelIP(vars["port"], arpOuterVlan, arpInnerVlan)
        str := fmt.Sprintf("trace: thrift: del_ip(%v, %v, %v) = (%v, %v)", vars["port"], outerVlan, innerVlan, retarp, err)
        log.Print(str)

        if err != nil || !retarp {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
            return
        }
    }

    proxyARPIP := net.ParseIP(attr.ProxyArpIP)
    arpIP := arp.Ip4T(IpToInt32(proxyARPIP))

    retarp, err := arpClient.AddIP(vars["port"], arpOuterVlan, arpInnerVlan, arpIP)
    str := fmt.Sprintf("trace: thrift: add_ip(%v, %v, %v, %v) = (%v, %v)", vars["port"], outerVlan, innerVlan, arpIP, retarp, err)
    log.Print(str)

    if err != nil || !retarp {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
        return
    }

    pt.Set(table, map[string]string{
        "vrf_id":       strconv.Itoa(attr.VrfID),
        "peer_ip":      attr.PeerIP,
        "proxy_arp_ip": attr.ProxyArpIP,
        "subnet":       attr.Subnet,
    }, "SET", "")

    if _, ok := configSnapshot.VrfMap[attr.VrfID]; ok {
        configSnapshot.VrfMap[attr.VrfID].QinQPortMap[table] = attr
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelDecapTunnelTypeDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    err := ValidateTunnelType(w, vars["tunnel_type"])
    if err != nil {
        return
    }
/*
    // Uncomment this code if we ever need to allow PA changes via REST API
    db := &conf_db_ops
    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VXLAN_TUNNEL", "default_vxlan_tunnel"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"tunnel_type"}, "")
        return
    }

    pt := swsscommon.NewProducerStateTable(db.swss_db, "VXLAN_TUNNEL")
    defer pt.Delete()
    pt.Del("default_vxlan_tunnel", "DEL", "")

    configSnapshot.DecapModel = TunnelDecapModel{}
*/

    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelDecapTunnelTypeGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops

    vars := mux.Vars(r)

    err := ValidateTunnelType(w, vars["tunnel_type"])
    if err != nil {
        return
    }

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VXLAN_TUNNEL", "default_vxlan_tunnel"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"tunnel_type"}, "")
        return
    }

    output := TunnelDecapReturnModel{
        TunnelType: vars["tunnel_type"],
        Attr: TunnelDecapModel{
            IPAddr: kv["src_ip"],
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigTunnelDecapTunnelTypePost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops

    vars := mux.Vars(r)

    err := ValidateTunnelType(w, vars["tunnel_type"])
    if err != nil {
        return
    }

    w.WriteHeader(http.StatusNoContent)
/* Comment out all the code below this point in this fn once Day 0 config for VTEP is complete */
    var attr TunnelDecapModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VXLAN_TUNNEL", "default_vxlan_tunnel"))
    if err != nil {
        /* WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "") */
        return
    }

    if kv != nil {
        /* WriteRequestError(w, http.StatusConflict, "0: Object already exists: Default Vxlan VTEP", []string{}, "") */
        return
    }

    pt := swsscommon.NewProducerStateTable(db.swss_db, "VXLAN_TUNNEL")
    defer pt.Delete()

    pt.Set("default_vxlan_tunnel", map[string]string{
        "src_ip": attr.IPAddr,
    }, "SET", "")

    configSnapshot.DecapModel = attr
}

func ConfigTunnelEncapVxlanGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    kv, err := SwssGetKVsMulti("TUNNEL_TABLE:encapsulation:vxlan:*")
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    tunnels := make([]TunnelReturnModel, 0, len(kv))

    for k, kvp := range kv {
        vnidStr := strings.Split(k, ":")[3]
        vnid, err := strconv.Atoi(vnidStr)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        vrfIDStr := kvp["vrf_id"]
        vrfID, err := strconv.Atoi(vrfIDStr)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }
        tunnels = append(tunnels, TunnelReturnModel{
            Vnid: vnid,
            Attr: TunnelModel{
                VrfID: vrfID,
            },
        })
    }

    WriteRequestResponse(w, tunnels, http.StatusOK)
}

func ConfigTunnelEncapVxlanVnidDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vnid, err := ValidateVnid(w, vars["vnid"])
    if err != nil {
        return
    }

    mseevnid := msee.MseeVniT(vnid)
    ret, err := mseeClient.UnmapVniToVrf(mseevnid)
    str := fmt.Sprintf("trace: thrift: unmap_vni_to_vrf(%v) = (%v, %v)", mseevnid, ret, err)
    
    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "TUNNEL_TABLE")
    defer pt.Delete()

    pt.Del("encapsulation:vxlan:"+vars["vnid"], "DEL", "")

    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelEncapVxlanVnidGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vnid, err := ValidateVnid(w, vars["vnid"])
    if err != nil {
        return
    }

    kv, err := SwssGetKVs("TUNNEL_TABLE:encapsulation:vxlan:" + vars["vnid"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }

    vrfID, err := ValidateVrfId(w, kv["vrf_id"])
    if err != nil {
        return
    }

    output := TunnelReturnModel{
        Vnid: vnid,
        Attr: TunnelModel{
            VrfID: vrfID,
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigTunnelEncapVxlanVnidPut(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vnid, err := ValidateVnid(w, vars["vnid"])
    if err != nil {
        return
    }

    var attr TunnelModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    _, err = CacheGetVrfName(attr.VrfID)
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vrf_id"}, "")
        return
    }

    mseevnid := msee.MseeVniT(vnid)
    mseeVrfID := msee.MseeVrfIDT(attr.VrfID)
    ret, err := mseeClient.MapVniToVrf(mseevnid, mseeVrfID)
    str := fmt.Sprintf("trace: thrift: map_vni_to_vrf(%v, %v) = (%v, %v)", mseevnid, mseeVrfID, ret, err)

    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "TUNNEL_TABLE")
    defer pt.Delete()

    pt.Set("encapsulation:vxlan:"+vars["vnid"], map[string]string{
        "vrf_id": strconv.Itoa(attr.VrfID),
    }, "SET", "")

    if _, ok := configSnapshot.VrfMap[attr.VrfID]; ok {
        configSnapshot.VrfMap[attr.VrfID].VxlanMap[vnid] = attr
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    kv, err := SwssGetKVsMulti("VROUTER_TABLE:*")
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    vrouters := make([]VirtualRouterReturnModel, 0, len(kv))

    for k, kvp := range kv {
        vrfIDStr := strings.Split(k, ":")[1]
        vrfID, err := strconv.Atoi(vrfIDStr)
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }

        vrfName := kvp["name"]
        vrouters = append(vrouters, VirtualRouterReturnModel{
            VrfID: vrfID,
            Attr: VirtualRouterModel{
                VrfName: vrfName,
            },
        })
    }

    WriteRequestResponse(w, vrouters, http.StatusOK)
}

func ConfigVrouterVrfIdDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    if vnet_id == 0 {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }
    vnet_id_str := strconv.FormatUint(uint64(vnet_id), 10)
    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VNET", vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error: GUID Cache and DB out of sync", []string{}, "")
        return
    }

    pt := swsscommon.NewProducerStateTable(db.swss_db, "VNET")
    defer pt.Delete()

    pt.Del(vnet_id_str, "DEL", "")
    CacheDeleteVnetGuidId(vars["vnet_name"])

    delete(configSnapshot.VrfMap, int(vnet_id))
    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    if vnet_id == 0 {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }
    vnet_id_str := strconv.FormatUint(uint64(vnet_id), 10)
    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VNET", vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, "")
        return
    }

    vnid, err := strconv.Atoi(kv["vni"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error, Non numeric vnid found in db", []string{}, "")
        return
    }

    output := VnetReturnModel{
        VnetName: vars["vnet_name"],
        Attr: VnetModel{
            Vnid: vnid,
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigVrouterVrfIdPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops

    vars := mux.Vars(r)
    if vars["vnet_name"] == "" {
        WriteRequestError(w, http.StatusNotFound, "VRF_ID/VNET_NAME cannot be NULL", []string{"tunnel_type"}, "")
        return
    }
    var attr VnetModel

    err := ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, "_VXLAN_TUNNEL", "default_vxlan_tunnel"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusConflict, "1: Default VxLAN VTEP must be created prior to creating VRF", []string{"tunnel_type"}, "")
        return
    }

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    if vnet_id != 0 {
        WriteRequestError(w, http.StatusConflict, "0: Object already exists: " + vars["vnet_name"], []string{}, "")
        return
    }

    vnet_id = CacheGenAndSetVnetGuidId(vars["vnet_name"])
    vnet_id_str := strconv.FormatUint(uint64(vnet_id), 10)

    kv, err = GetKVs(db.db_num, generateDBTableKey(db.separator, "_VNET", vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error: GUID Cache and DB out of sync", []string{}, "")
        return
    }

    pt := swsscommon.NewProducerStateTable(db.swss_db, "VNET")
    defer pt.Delete()

    pt.Set(vnet_id_str, map[string]string{
        "vxlan_tunnel": "default_vxlan_tunnel",
        "vni": strconv.Itoa(attr.Vnid),
        "guid": vars["vnet_name"],
    }, "SET", "")

    configSnapshot.VrfMap[int(vnet_id)] = VrfSnapshotModel{
        VrfInfo:     attr,
        VxlanMap:    make(map[int]TunnelModel),
        PortMap:     make(map[string]PortModel),
        QinQPortMap: make(map[string]QInQModel),
        RoutesMap:   make(map[string]RouteModel),
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdRoutesDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vrfID, err := ValidateVrfId(w, vars["vrf_id"])
    if err != nil {
        return
    }

    _, err = CacheGetVrfName(vrfID)
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vrf_id"}, "")
        return
    }

    mseeVrfID := msee.MseeVrfIDT(vrfID)

    vnidMatch := -1
    if len(r.URL.Query()["vnid"]) == 1 {
        vnid := r.URL.Query()["vnid"][0]
        vnidMatch, err = ValidateVnid(w, vnid)
        if err != nil {
            return
        }
    } else if len(r.URL.Query()["vnid"]) > 1 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnid"}, "May only specify one vnid")
        return
    }

    var attr []RouteModel

    if (*r).ContentLength > 0 {
        err = ReadJSONBody(w, r, &attr)
        if err != nil {
            // The error is already handled in this case
            return
        }
    } else {
        attr, err = SwssGetVrouterRoutes(vrfID, vnidMatch, "*")
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
            return
        }
    }

    var failed []RouteModel
    var removed []RouteModel

    pt := swsscommon.NewProducerStateTable(swssDB, "VROUTER_ROUTES_TABLE")
    defer pt.Delete()

    for _, r := range attr {
        table := vars["vrf_id"] + ":" + r.IPPrefix
        kv, err := SwssGetKVs("VROUTER_ROUTES_TABLE:" + table)
        if err != nil || kv == nil {
            failed = append(failed, r)
            continue
        }

        mseeDstIPPrefix := GetThriftIPPrefix(r.IPPrefix)
        if (r.NextHopType == "vxlan-tunnel") {
            ret, err := mseeClient.DeleteEncapRoute(mseeVrfID, &mseeDstIPPrefix)
            str := fmt.Sprintf("trace: thrift: delete_encap_route(%v, %v) = (%v, %v)", mseeVrfID, mseeDstIPPrefix, ret, err)

            if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
                return
            }
        } else if (r.NextHopType == "ip") {
            ret, err := mseeClient.DeleteDecapRoute(mseeVrfID, &mseeDstIPPrefix)
            str := fmt.Sprintf("trace: thrift: delete_decap_route(%v, %v) = (%v, %v)", mseeVrfID, mseeDstIPPrefix, ret, err)

            if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
                return
            }
        } else {
            failed = append(failed, r)
            r.Error = fmt.Sprintf("NextHopType is not supported %v", r.NextHopType)
            log.Printf("warning: %v", r.Error)
            continue
        }   

        pt.Del(table, "DEL", "")
        removed = append(removed, r)

        if _, ok := configSnapshot.VrfMap[vrfID]; ok {
            delete(configSnapshot.VrfMap[vrfID].RoutesMap, r.IPPrefix)
        }
    }

    output := RouteDeleteReturnModel{
        Removed: removed,
        Failed:  failed,
    }

    WriteRequestResponse(w, output, http.StatusCreated)
}

func ConfigVrouterVrfIdRoutesGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vrfID, err := ValidateVrfId(w, vars["vrf_id"])
    if err != nil {
        return
    }

    _, err = CacheGetVrfName(vrfID)
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vrf_id"}, "")
        return
    }

    ipprefix := "*"
    if len(r.URL.Query()["ip_prefix"]) == 1 {
        ipprefix = r.URL.Query()["ip_prefix"][0]
        _, _, err := ParseIPPrefix(ipprefix)
        if err != nil {
            WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ip_prefix"}, "Invalid ip_prefix")
            return
        }
    } else if len(r.URL.Query()["ip_prefix"]) > 1 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ip_prefix"}, "May only specify one ip_prefix")
        return
    }

    vnidMatch := -1
    if len(r.URL.Query()["vnid"]) == 1 {
        vnid := r.URL.Query()["vnid"][0]
        vnidMatch, err = ValidateVnid(w, vnid)
        if err != nil {
            return
        }
    } else if len(r.URL.Query()["vnid"]) > 1 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnid"}, "May only specify one vnid")
        return
    }

    routes, err := SwssGetVrouterRoutes(vrfID, vnidMatch, ipprefix)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    WriteRequestResponse(w, routes, http.StatusOK)
}

func ConfigVrouterVrfIdRoutesPut(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    vrfID, err := ValidateVrfId(w, vars["vrf_id"])
    if err != nil {
        return
    }

    mseeVrfID := msee.MseeVrfIDT(vrfID)

    var attr []RouteModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    _, err = CacheGetVrfName(vrfID)
    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vrf_id"}, "")
        return
    }

    pt := swsscommon.NewProducerStateTable(swssDB, "VROUTER_ROUTES_TABLE")
    defer pt.Delete()

    var added []RouteModel
    var failed []RouteModel
    var updated []RouteModel

    var arpRequest []*arp.ReqTuplesT

    var i int32
    i = -1

    for _, r := range attr {
        i++
        if r.NextHopType == "ip" {
            arpTuples, err := SwssGetVrfPorts(vars["vrf_id"])
            if err != nil || len(arpTuples) == 0 {
                continue
            }

            arpRequest = append(arpRequest, &arp.ReqTuplesT{
                IP: arp.Ip4T(IpToInt32(net.ParseIP(r.NextHop))),
                Index: i,
                Tuples: arpTuples,
            })

        }
    }

    var arpResponse []*arp.RepTupleT
    if len(arpRequest) > 0 {
        var err error
        arpResponse, err = arpClient.RequestMac(arpRequest)
        str := fmt.Sprintf("trace: thrift: request_mac(%v) = (%v, %v)", arpRequest, arpResponse, err)
        log.Print(str)

        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
            return
        }
    }

    i = -1

    for _, r := range attr {
        i++

        _, exist, err := SwssGetVrouterRoute(vrfID, r.IPPrefix)
        if err != nil {
            r.Error = "Internal service error"
            failed = append(failed, r)
            continue
        }

        mseeDstIPPrefix := GetThriftIPPrefix(r.IPPrefix)
        mseeDstIP := GetThriftIPAddress(net.ParseIP(r.NextHop))

        kv := map[string]string{
            "nexthop_type": r.NextHopType,
            "nexthop":      r.NextHop,
        }

        if kv["nexthop_type"] == "vxlan-tunnel" {
            kv["vxlanid"] = strconv.Itoa(r.Vnid)
            kv["mac_address"] = r.MACAddress
            kv["port"] = r.Port

            if r.SrcIP != "" {
                kv["src_ip"] = r.SrcIP
            }

            vnidkv, err := SwssGetKVs("TUNNEL_TABLE:encapsulation:vxlan:"+kv["vxlanid"])
            if err != nil || vnidkv == nil {
                failed = append(failed, r)
                continue
            }

            mseeVni := msee.MseeVniT(r.Vnid)
            mac, _ := net.ParseMAC(r.MACAddress)
            mseeMACAddress := msee.MseeMacT(MacToInt64(mac))

            vxlanPort := uint16(65330)
            if r.Port == "standard" {
                vxlanPort = uint16(4789)
            }

            mseeUDPPort := msee.MseeUDPPortT(vxlanPort)

            ret, err := mseeClient.AddEncapRoute(mseeVrfID, &mseeDstIPPrefix, &mseeDstIP, mseeMACAddress, mseeVni, mseeUDPPort)
            str := fmt.Sprintf("trace: thrift: add_encap_route(%v, %v, %v, %v, %v, %v) = (%v, %v)", mseeVrfID, mseeDstIPPrefix, mseeDstIP, mseeMACAddress, mseeVni, mseeUDPPort, ret, err)

            if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
                 return
            }
        } else if kv["nexthop_type"] == "ip" {
            for j, rep := range arpResponse {
                if rep.Index == i {
                    // Remove this element
                    arpResponse = append(arpResponse[:j], arpResponse[j+1:]...)

                    if !rep.IsFound {
                        r.Error = "Nexthop not found"
                        break
                    }

                    kv["mac_address"] = net.HardwareAddr(rep.Mac).String()

                    mseeMACAddress := msee.MseeMacT(MacToInt64(net.HardwareAddr(rep.Mac)))
                    portID, err := PortToPortID(rep.Request.IfaceName)
                    if err != nil {
                        r.Error = "Internal service error"
                        break
                    }

                    mseePort := msee.MseePortT(portID)
                    mseeStag := msee.MseeVlanT(rep.Request.Stag)
                    mseeCtag := msee.MseeVlanT(rep.Request.Ctag)

                    ret, err := mseeClient.AddDecapRoute(mseeVrfID, &mseeDstIPPrefix, mseeMACAddress, mseePort, mseeStag, mseeCtag)
                    str := fmt.Sprintf("trace: thrift: add_decap_route(%v, %v, %v, %v, %v, %v) = (%v, %v)",
                        mseeVrfID, mseeDstIPPrefix, mseeMACAddress, mseePort, mseeStag, mseeCtag, ret, err)

                    if WriteRequestErrorForMSEEThrift(w, err, ret, str) {
                        return
                    }

                    goto success
                }
            }

            failed = append(failed, r)
            continue
        } else {
            r.Error = "Not implemented"
            failed = append(failed, r)
            continue
        }

success:
        pt.Set(vars["vrf_id"]+":"+r.IPPrefix, kv, "SET", "")

        if !exist {
            added = append(added, r)
        } else {
            updated = append(updated, r)
        }

        if _, ok := configSnapshot.VrfMap[vrfID]; ok {
            configSnapshot.VrfMap[vrfID].RoutesMap[r.IPPrefix] = r
        }
    }

    output := RoutePutReturnModel{
        Added:   added,
        Updated: updated,
        Failed:  failed,
    }

    WriteRequestResponse(w, output, http.StatusCreated)
}

func StateInterfaceGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    ns, err := GetAllNetworkStatuses()

    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    var output []interface{}
    for port, status := range ns {
        var statusStr string
        if status {
            statusStr = "up"
        } else {
            statusStr = "down"
        }

        output = append(output, InterfaceReturnModel{
            Port: port,
            Attr: InterfaceModel{
                AdminState: statusStr,
            },
        })
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func StateInterfacePortGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    ns, err := GetNetworkStatus(vars["port"])

    if err != nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"port"}, "")
        return
    }

    var statusStr string
    if ns {
        statusStr = "up"
    } else {
        statusStr = "down"
    }

    output := InterfaceReturnModel{
        Port: vars["port"],
        Attr: InterfaceModel{
            AdminState: statusStr,
        },
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func GetSwitchCounterFromCounterDB() (groupcounters map[msee.MseeCounterName]int64 , err error) {

    groupcounters = make(map[msee.MseeCounterName]int64)

    for port, counterID := range portCounterIDMap {
        if _, ok := portIDMap[port]; !ok {
            // The port is not a port for customer devices
            continue
        }

        var kv map[string]string
        kv, err = CounterGetKVs("COUNTERS:" + counterID)

        if err != nil {
            return
        }

        if kv != nil {
            for k, v := range kv {
                value, err := strconv.ParseInt(v, 10, 64)           

                if err != nil {
                    log.Printf("error: can not counter to integer for %v in counter %v: %v. Skip the counter.", port, k, v)
                    continue
                }

                groupcounters[msee.MseeCounterName(port + "." + k)] = value
            }
        }
    }

    return
}

func StateCounterGetHelper(group msee.MseeGroupT) (counters map[msee.MseeGroupT]map[msee.MseeCounterName]int64 , err error) {

    counters, err = mseeClient.GetCounters(group)
    str := fmt.Sprintf("trace: thrift: get_counters(%v) = (%v, %v)", group, counters, err)
    log.Print(str)

    if err != nil {
        err = errors.New(str)
        return
    }

    for groupk, groupv := range counters {
        if groupk == "dpdk.switch_ports" {
            re := regexp.MustCompile("^[0-9]+[.]")
            newGroupCounters := make(map[msee.MseeCounterName]int64)

            for counterk, counterv := range groupv {
                portIDString := re.FindString(string(counterk))

                if len(portIDString) == 0 {
                    log.Printf("error: can not find port id in counter %v: %v. Skip the counter.", counterk, counterv)
                    continue
                }
                portIDString = portIDString[:len(portIDString) - 1]

                portID, err := strconv.Atoi(portIDString)
                if err != nil {
                    log.Printf("error: can not convert port id to integer for %v in counter %v: %v. Skip the counter.", portIDString, counterk, counterv)
                    continue
                }

                portName, err := PorIDToPort(portID)
                if err != nil {
                    log.Printf("error: can not find port name for %v in counter %v: %v. Skip the counter.", portID, counterk, counterv)
                    continue
                }

                newCounterk := re.ReplaceAllString(string(counterk), portName + ".")

                newGroupCounters[msee.MseeCounterName(newCounterk)] = counterv
            }

            counters["dpdk.switch_ports"] = newGroupCounters
            break
        }
    }

    return counters, nil
}

func StateCounterGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

        group := msee.MseeGroupT("")
        counters, err:= StateCounterGetHelper(group)

        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, err.Error())
            return
        }

        groupcounters, err:= GetSwitchCounterFromCounterDB()

        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, err.Error())
            return
        }

        counters["switch.switch_ports"] = groupcounters

        WriteRequestResponse(w, counters, http.StatusOK)
}

func StateCounterGroupGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    err := ValidateCounterGroupName(w, vars["group"])
    if err != nil {
        return
    }

    counters :=  make(map[msee.MseeGroupT]map[msee.MseeCounterName]int64)

    if (vars["group"] == "switch.switch_ports") {
        groupcounters, err:= GetSwitchCounterFromCounterDB()

        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, err.Error())
            return
        }

        counters["switch.switch_ports"] = groupcounters
    } else {
        group := msee.MseeGroupT(vars["group"])
        counters , err = StateCounterGetHelper(group)

        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, err.Error())
            return
        }
    }

    WriteRequestResponse(w, counters, http.StatusOK)
}

func StateHistogramGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    hist, err := mseeClient.GetHist()
    str := fmt.Sprintf("trace: thrift: get_hists() = (%v, %v)", hist, err)
    log.Print(str)

    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
        return
    }

    WriteRequestResponse(w, hist, http.StatusOK)
}

func StateStatisticsGetHelper(w http.ResponseWriter, group msee.MseeGroupT) {

    statistics, err := mseeClient.GetStatistics(group)
    str := fmt.Sprintf("trace: thrift: get_statistics(%v) = (%v, %v)", group, statistics, err)
    log.Print(str)

    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, str)
        return
    }

    WriteRequestResponse(w, statistics, http.StatusOK)
}

func StateStatisticsGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    group := msee.MseeGroupT("")

    StateStatisticsGetHelper(w, group)
}

func StateStatisticsGroupGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    vars := mux.Vars(r)

    err := ValidateStatisticsGroupName(w, vars["group"])
    if err != nil {
        return
    }

    group := msee.MseeGroupT(vars["group"])

    StateStatisticsGetHelper(w, group)
}


