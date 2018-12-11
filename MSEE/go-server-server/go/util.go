package mseeserver

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "log"
    "mseethrift"
    "net"
    "net/http"
    "strconv"
)

func WriteRequestErrorForMSEEThrift(w http.ResponseWriter, err error, r msee.ResultT, details string) bool {
    log.Print(details)
    
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, details)
        return true
    } else if r == msee.ResultT_INVALID_PARAMETERS {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{}, details)
        return true
    } else if r == msee.ResultT_NO_MEMORY {
        WriteRequestError(w, http.StatusForbidden, "Capacity insufficient", []string{}, details)
        return true
    } else if r == msee.ResultT_ALREADY_EXISTS {
        WriteRequestError(w, http.StatusMethodNotAllowed, "Object already exists", []string{}, details)
        return true
    } else if r == msee.ResultT_NOT_FOUND {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{}, details)
        return true
    }

    return false
}

func WriteRequestError(w http.ResponseWriter, code int, message string, fields []string, details string) {
    e := ErrorInner{
        Code:    code,
        Message: message,
        Fields:  fields,
        Details: details,
    }

    b, err := json.Marshal(ErrorModel{e})
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        b = []byte(`
{
  "error": {
    "code": 500,
    "message": "Internal service error"
  }
}`)
    } else {
        w.WriteHeader(code)
    }

    log.Printf(
        "error: Request ends with error message %s",
        b,
    )

    w.Write(b)
}

func WriteJsonError(w http.ResponseWriter, err error) {
    switch t := err.(type) {
    case *json.SyntaxError:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{}, "Invalid character in JSON")
    case *json.UnmarshalTypeError:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{t.Field}, "JSON field does not match required type")
    case *MissingValueError:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{t.Field}, "Missing JSON field")
    case *InvalidFormatError:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{t.Field}, t.Message)
    default:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{}, "Failed to decode JSON")
    }
}

func WriteRequestResponse(w http.ResponseWriter, jsonObj interface{}, code int) {
    b, err := json.Marshal(jsonObj)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
    } else {
        w.WriteHeader(code)
        w.Write(b)
    }
}

func ReadJSONBody(w http.ResponseWriter, r *http.Request, attr interface{}) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return err
    }

    log.Printf(
        "debug: request: body: %s",
        body,
    )

    err = json.Unmarshal(body, attr)
    if err != nil {
        WriteJsonError(w, err)
        return err
    }

    return nil
}

func IsValidIP(ipstr string) bool {
    ip := net.ParseIP(ipstr)
    return (ip != nil) && (ip.To4() != nil)
}

func IsValidIPBoth(ipstr string) bool {
    ip := net.ParseIP(ipstr)
    return ip != nil
}

func ParseIPBothPrefix(ipprefix string) (ipstr string, length int, err error) {
    ip, net, err := net.ParseCIDR(ipprefix)
    if err != nil {
        return
    }

    ipstr = ip.String()
    length, _ = net.Mask.Size()

    return
}

func ParseIPPrefix(ipprefix string) (ipstr string, length int, err error) {
    ip, net, err := net.ParseCIDR(ipprefix)
    if err != nil {
        return
    }

    if ip.To4() == nil {
        err = errors.New("Only IPv4 supported")
        return
    }

    ipstr = ip.String()
    length, _ = net.Mask.Size()

    return
}

func ValidateVnid(w http.ResponseWriter, vnidStr string) (vnid int, err error) {
    vnid, err = strconv.Atoi(vnidStr)
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnid"}, "vnid must be an integer")
        return
    }

    if vnid >= 0x1000000 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnid"}, "vnid must be < 2^24")
        err = errors.New("vnid must be < 2^24")
        return
    }
    return
}

func ValidateVrfId(w http.ResponseWriter, vrfIdStr string) (vrfId int, err error) {
    vrfId, err = strconv.Atoi(vrfIdStr)
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vrf_id"}, "vrf_id must be an integer")
        return
    }
    return
}

func ValidateTunnelType(w http.ResponseWriter, tunnelType string) error {
    if tunnelType != "vxlan" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"tunnel_type"}, "Only tunnel_type==vxlan supported")
        return errors.New("tunnel_type must be vxlan")
    }
    return nil
}

func ValidateCounterGroupName(w http.ResponseWriter, groupName string) error {
    if groupName != "dpdk.total" && groupName != "dpdk.switch_ports" && groupName != "dpdk.nic" && groupName != "switch.switch_ports" && groupName != "dpdk.switch_vnis" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"group"}, "group name must be dpdk.total, dpdk.switch_ports, dpdk.nic, switch.switch_ports or dpdk.switch_vnis")
        return errors.New("group name must be dpdk.total, dpdk.switch_ports, dpdk.nic, switch.switch_ports or dpdk.switch_vnis")
    }
    return nil
}

func ValidateStatisticsGroupName(w http.ResponseWriter, groupName string) error {
    if groupName != "rings" && groupName != "mempools" && groupName != "fibs" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"group"}, "group name must be rings, mempools or fibs")
        return errors.New("group name must be rings, mempools or fib")
    }
    return nil
}

func ValidateStagCtag(w http.ResponseWriter, stagStr string, ctagStr string) (stag int, ctag int, err error) {
    stag, err = strconv.Atoi(stagStr)
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"stag"}, "stag must be an integer")
        return
    }

    if stag <= 1 || stag >= 4096 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"stag"}, "stag must be an integer between 2 and 4095")
        err = errors.New("stag must be between 2 and 4095")
        return
    }

    ctag, err = strconv.Atoi(ctagStr)
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ctag"}, "ctag must be an integer")
        return
    }

    if ctag <= 1 || ctag >= 4096 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ctag"}, "ctag must be an integer between 2 and 4095")
        err = errors.New("ctag must be between 2 and 4095")
        return
    }

    return
}

func IpToInt32(ipAddr net.IP) (int32) {
    addr := ipAddr.To4()
    return int32(addr[0]) << 24 | int32(addr[1]) << 16 | int32(addr[2]) << 8 | int32(addr[3])
}

func MacToInt64(macAddr net.HardwareAddr) (int64) {
    addr := macAddr
    return int64(addr[0]) << 40 | int64(addr[1]) << 32 | int64(addr[2]) << 24 | int64(addr[3]) << 16 | int64(addr[4]) << 8 | int64(addr[5])
}

func PortToPortID(port string) (portID int, err error) {
    portID, exist := portIDMap[port]
    if !exist {
        err = errors.New("port specified does not exist")
    }
    return
}

func PorIDToPort(portID int) (port string, err error) {
    port, exist := portNameMap[portID]
    if !exist {
        err = errors.New("Specified port_id doesn't exist")
    }
    return
}

func PortToVlanID(port string) (vlanID int, err error) {
    vlanID, err = PortToPortID(port)
    vlanID += *VlanStartFlag
    return
}

func GetThriftIPPrefix(ipPrefix string) (ret msee.MseeIPPrefixT) {
    _, net, _ := net.ParseCIDR(ipPrefix)
    ipHelper := GetThriftIPAddress(net.IP)
    ret.IP = &ipHelper
    len, _ := net.Mask.Size()
    ret.MaskLength = msee.MseePrefixLenT(len)
    return
}

func GetThriftIPAddress(ip net.IP) (ret msee.MseeIPAddressT) {
    ret.IP = new(msee.MseeIPT)
    if ip.To4() != nil {
        ip4 := msee.MseeIp4T(IpToInt32(ip))
        ret.IP.Ip4 = &ip4
        ret.Type = msee.IPTypeT_v4
    } else {
        // TODO: Implement IPv6 conversion
        log.Printf("debug: IPv6 conversion not supported in GetThriftIPAddress")
        var ip6 msee.MseeIp6T
        ret.IP.Ip6 = &ip6
        ret.Type = msee.IPTypeT_v6
    }
    return
}

func validateVlanID(vlan_id_str string) (vlanID int, err error) {
   vlanID, err = strconv.Atoi(vlan_id_str)
   if err == nil {
       if vlanID < 2 || vlanID > 4094 {
           err = errors.New("vlanID out of range " + vlan_id_str)
       }
   }
   return
}
