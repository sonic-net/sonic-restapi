package restapi

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
    "regexp"
)

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


func WriteRequestErrorWithSubCode(w http.ResponseWriter, code int, sub_code int, message string, fields []string, details string) {
    e := ErrorInner{
        Code:     code,
        SubCode: &sub_code,
        Message:  message,
        Fields:   fields,
        Details:  details,
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

func isV4orV6(ipaddr string) (int) {
    ip := net.ParseIP(ipaddr)
    if ip.To4() != nil {
        return 4
    }
    return 6
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

func ExtractIPPrefixFromKey(key string, dbseperator string) (ipprefix string, err error) {
    ctr := 0
	var idx int
	for i, c := range key {
		if strings.Compare(string(c), dbseperator) == 0 {
			ctr += 1
		}
		if ctr == 2 {
			idx = i
			break
		}
	}
	ipprefix = key[idx+1 : len(key)]
    _, _, err = net.ParseCIDR(ipprefix)
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

func ValidateTunnelType(w http.ResponseWriter, tunnelType string) error {
    if tunnelType != "vxlan" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"tunnel_type"}, "Only tunnel_type==vxlan supported")
        return errors.New("tunnel_type must be vxlan")
    }
    return nil
}

func IpToInt32(ipAddr net.IP) (int32) {
    addr := ipAddr.To4()
    return int32(addr[0]) << 24 | int32(addr[1]) << 16 | int32(addr[2]) << 8 | int32(addr[3])
}

func MacToInt64(macAddr net.HardwareAddr) (int64) {
    addr := macAddr
    return int64(addr[0]) << 40 | int64(addr[1]) << 32 | int64(addr[2]) << 24 | int64(addr[3]) << 16 | int64(addr[4]) << 8 | int64(addr[5])
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

func generateVlanPrefixInVnet(vnet_id_str string) (vlanPrefixArr []string, err error) {
    db := &app_db_ops
    var rt_tb_key string
    rt_tb_name := LOCAL_ROUTE_TB
    if *RunApiAsLocalTestDocker {
        rt_tb_name = "_"+LOCAL_ROUTE_TB
    }
    rt_tb_key = generateDBTableKey(db.separator, rt_tb_name, vnet_id_str, "*")
    kv, err := GetKVsMulti(db.db_num, rt_tb_key)
    if err != nil {
        return vlanPrefixArr, err
    }
    for k, _ := range kv {
        ipprefix := strings.Split(k, db.separator)[2]
        vlanPrefixArr = append(vlanPrefixArr, ipprefix)
    }
    return
}

func isBMNextHop(ipprefix string, vlanPrefixArr []string) (bm_next_hop bool, err error) {
    bm_next_hop = false
    if (vlanPrefixArr == nil || len(vlanPrefixArr) == 0) {
        return
    }
    ip, _, err := net.ParseCIDR(ipprefix)
    if err != nil {
        return bm_next_hop, err
    }
	 if (ip.To4() == nil && strings.HasSuffix(ipprefix, "/128")) || (ip.To4() != nil && strings.HasSuffix(ipprefix, "/32"))  {
        for _, vlan_prefix := range vlanPrefixArr {
             vlan_ip, vlan_netw, err := net.ParseCIDR(vlan_prefix)
		       if err != nil {
		           return bm_next_hop, err
		       }
				 if ((ip.To4 == nil && vlan_ip.To4 == nil) || (ip.To4 != nil && vlan_ip.To4 != nil)) && vlan_netw.Contains(ip) {
				     bm_next_hop = true
                 return bm_next_hop, err
			    }
		  }
    }
	 return
}

func isLocalTunnelNexthop(ipNextHop string) (local_next_hop bool) {
    local_next_hop = false
    lpbk_ips := localTunnelLpbkIps
    if (ipNextHop == "" || len(lpbk_ips) == 0) {
        return
    }

    for _, s := range lpbk_ips {
        if s == ipNextHop {
            local_next_hop = true
            return local_next_hop
        }
    }
    return
}

func vlan_dependencies_exist(vlan_name string) (vlan_dep bool, err error) {
    db := &conf_db_ops
    vlan_dep = false
    neigh_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_NEIGH_TB, vlan_name, "*"))
    if err != nil {
        return
    }
    vlan_mem_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name, "*"))
    if err != nil {
        return
    }
    if len(neigh_kv) > 0 || len(vlan_mem_kv) > 0 {
        vlan_dep = true
    }
    return
}


func vnet_dependencies_exist(vnet_id_str string) (vnet_dep bool, err error) {
     db := &app_db_ops
     var rt_tb_key string
     vnet_dep = false

     rt_tb_name := ROUTE_TUN_TB
     if *RunApiAsLocalTestDocker {
         rt_tb_name = "_"+ROUTE_TUN_TB
     }
     rt_tb_key = generateDBTableKey(db.separator, rt_tb_name, vnet_id_str, "*")
     routes_kv, err := GetKVsMulti(db.db_num, rt_tb_key)/* generateDBTableKey(db.separator, ROUTE_TUN_TB, vnet_id_str, "*"))*/
     if err != nil {
        return
     } else if len(routes_kv) > 0 {
        vnet_dep = true
        return
     }
     vlan_if_kv, err := GetKVsMulti(conf_db_ops.db_num, generateDBTableKey(conf_db_ops.separator, VLAN_INTF_TB,"*"))
     if err != nil {
        return
     }
     for _,v := range vlan_if_kv {
        if v["vnet_name"] == vnet_id_str {
            vnet_dep = true
            return
        }
     }
     return
}

func vlan_validator(w http.ResponseWriter, vlan_id_str string) (vlan_id int, err error) {
    db := &conf_db_ops
    vlan_id, err = validateVlanID(vlan_id_str)
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vlan_id"}, "")
        return vlan_id, err
    }
    vlan_name := VLAN_NAME_PREF + vlan_id_str
    vlan_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_TB, vlan_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{"vlan_id"}, "")
        return vlan_id, errors.New("Internal service err")
    }
    if vlan_kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"vlan_id"}, "")
        return vlan_id, errors.New("Vlan obj not found")
    }
    return
}

func get_and_validate_vnet_id(w http.ResponseWriter, vnet_name string) (vnet_id_str string, kv map[string]string, err error) {
    db := &conf_db_ops
    vnet_id := CacheGetVnetGuidId(vnet_name)
    if vnet_id == 0 {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{vnet_name}, "")
        err = errors.New("Vnet obj not found")
        return
    }
    vnet_id_str = VNET_NAME_PREF + strconv.FormatUint(uint64(vnet_id), 10)
    kv, err = GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error: GUID Cache and DB out of sync", []string{}, "")
        return
    }
    return
}

func parse_ping_output(ping_output string) (parsed_output PingReturnModel) {
    var split_string []string
    split_string = strings.Split(ping_output, " ")
    for index, el := range split_string {
	if el == "transmitted," {
	    re := regexp.MustCompile("[0-9]")
	    parsed_output.PacketsTransmitted = re.FindString(split_string[index-2])
	}
	if el == "received," {
		parsed_output.PacketsReceived = split_string[index-1]
	}
	if el == "min/avg/max/mdev" {
		parsed_output.MinRTT = strings.Split(split_string[index+2], "/")[0]
		parsed_output.MaxRTT = strings.Split(split_string[index+2], "/")[2]
		parsed_output.AvgRTT = strings.Split(split_string[index+2], "/")[1]

	}

    }
    return parsed_output
}
