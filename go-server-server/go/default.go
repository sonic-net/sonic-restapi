package restapi

import (
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
    "swsscommon"
    "time"
    "github.com/gorilla/mux"
    "os/exec"
)

const RESRC_EXISTS int = 0
const DEP_MISSING int  = 1
const DELETE_DEP  int  = 2
const DEFAULT_PING_COUNT_STR string = "4"
const PING_COMMAND_STR string = "ping"

func StateHeartbeatGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var availableRoutes int = -1
    db := &ctr_db_ops
    crm_stats_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, CRM_TB, "STATS"))
    if err != nil {
        log.Printf("Fetching CRM:STATS key from Counters DB failed")
    } else {
        availableRoutes, _ = strconv.Atoi(crm_stats_kv["crm_stats_ipv4_route_available"])
    }

    output := HeartbeatReturnModel{
        ServerVersion: ServerAPIVersion,
        ResetGUID: ServerResetGuid,
        ResetTime: ServerResetTime,
        RoutesAvailable: availableRoutes,
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigResetStatusGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var output ConfigResetStatusModel
    if ConfigResetStatus == true {
        output.ResetStatus = "true"
    } else {
        output.ResetStatus = "false"
    }
    WriteRequestResponse(w, output, http.StatusOK)    
}

func ConfigResetStatusPost(w http.ResponseWriter, r *http.Request) {
    var attr ConfigResetStatusModel
    
    ReadJSONBody(w, r, &attr)
    switch attr.ResetStatus {
    case "true":
        ConfigResetStatus = true
    case "false":
        ConfigResetStatus = false
    default:
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"reset_status"}, "only true/false values accepted")
        return
    }
    CacheSetResetStatusInfo(ConfigResetStatus)
    ConfigResetStatusGet(w, r)    
}

func ConfigBgpProfilePost(w http.ResponseWriter, r *http.Request) {
    var attr BgpProfileModel
    vars := mux.Vars(r)
    db := &app_db_ops

    ReadJSONBody(w, r, &attr)

    bgp_profile_t := swsscommon.NewTable(db.swss_db, BGP_PROFILE_TABLE)
    defer bgp_profile_t.Delete()

    bgp_profile_t.Set(vars["profile_name"], map[string]string {
                        "community_id": attr.CommunityId,
        }, "SET", "")
    w.WriteHeader(http.StatusNoContent)   
}

func ConfigBgpProfileDelete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    db := &app_db_ops

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, BGP_PROFILE_TABLE, vars["profile_name"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv["community_id"] == "" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"profile_name"}, "Invalid profile_name")
        return
    }

    bgp_profile_t := swsscommon.NewTable(db.swss_db, BGP_PROFILE_TABLE)
    defer bgp_profile_t.Delete()

    bgp_profile_t.Del(vars["profile_name"], "DEL", "")
    w.WriteHeader(http.StatusNoContent)   
}

func ConfigBgpProfileGet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    db := &app_db_ops

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, BGP_PROFILE_TABLE, vars["profile_name"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv["community_id"] == "" {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"profile_name"}, "Invalid profile_name")
        return
    }

    output := BgpProfileModel {
        CommunityId: kv["community_id"],
    }
    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfaceVlanGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)
    var attr VlanModel

    vlan_id, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    vlan_if_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_if_kv != nil {
            vnet_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vlan_if_kv["vnet_name"]))
            if err != nil || vnet_kv == nil {
                 WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
                 return
            }
            attr.Vnet_id = vnet_kv["guid"]
    }

    vlan_pref_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name, "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if len(vlan_pref_kv) == 1 {
        for k,_ := range vlan_pref_kv {
            table_key := k[(len(generateDBTableKey(db.separator,VLAN_INTF_TB, vlan_name)) + 1):]
            attr.IPPrefix = table_key
        }
    } else if len(vlan_pref_kv) > 1 {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    output := VlanReturnModel{
        VlanID: vlan_id,
        Attr: attr,
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfaceVlanDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    _, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    vlan_if_pt := swsscommon.NewTable(db.swss_db, VLAN_INTF_TB)
    defer vlan_if_pt.Delete()

    vlan_dep, err := vlan_dependencies_exist(vlan_name)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_dep {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, DELETE_DEP,
             "Deleting object that has child dependency, child element must be deleted first", []string{}, "")
        return
    }

    vlan_pref_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name, "*"))
    if err != nil ||  len(vlan_pref_kv) > 1 {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    vlan_if_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    /* Delete sequence: 1. local subnet route(1s delay) 2. Vlan Interface IP prefix table(1s delay), 3. Vlan Interface table, 4. Vlan */
    /* Delete 1 */
    if len(vlan_pref_kv) == 1 && vlan_if_kv != nil {
        for k,_ := range vlan_pref_kv {
            ip_pref := k[(len(generateDBTableKey(db.separator,VLAN_INTF_TB, vlan_name)) + 1):]
             _, vlan_netw, _ := net.ParseCIDR(ip_pref)
            vnet_id_str := vlan_if_kv["vnet_name"]
            local_subnet_route_pt := swsscommon.NewProducerStateTable(app_db_ops.swss_db, LOCAL_ROUTE_TB)
            defer local_subnet_route_pt.Delete()
            local_subnet_route_pt.Del(generateDBTableKey(app_db_ops.separator, vnet_id_str, vlan_netw.String()), "DEL", "")
        }
    }

    /* Delete 2 */
    if len(vlan_pref_kv) == 1 {
        /* Sleep as we deleted local subnet route */
        time.Sleep(time.Second)
        for k,_ := range vlan_pref_kv {
            table_key := k[len(VLAN_INTF_TB)+ 1:]
            vlan_if_pt.Del(table_key, "DEL", "")
        }
    }

    /* Delete 3 */
    if vlan_if_kv != nil {
        if len(vlan_pref_kv) == 1 {
            /* Sleep only if we previously deleted the VLAN_INTERFACE ip_prefix table */
            time.Sleep(time.Second)
        }
        vlan_if_pt.Del(vlan_name, "DEL", "")
    }

    /* Delete 4 */
    pt := swsscommon.NewTable(db.swss_db, VLAN_TB)
    defer pt.Delete()
    pt.Del(vlan_name, "DEL", "")

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlanPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)
    var vnet_id uint32
    var vnet_id_str string

    _, err := validateVlanID(vars["vlan_id"])
    if err != nil {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vlan_id"}, "")
        return
    }

    var attr VlanModel
    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    /* Config validation and failure reporting */
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]
    vlan_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_TB, vlan_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_kv != nil {
        vlan_if_kv, _ := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name))
        if vlan_if_kv != nil {
            if vnet_name, ok := vlan_if_kv["vnet_name"]; ok {
                vnet_kv, _ := GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vnet_name))
                WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
                    "Object already exists: {\"vlan_name\":\"" + vlan_name + "\", \"vnet_id\":\"" + 
                    vnet_kv["guid"] +"\"}", []string{}, "")
                return
            }
        }
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
            "Object already exists: " + vlan_name, []string{}, "")
        return
    }

    if attr.Vnet_id != "" {
        vnet_id = CacheGetVnetGuidId(attr.Vnet_id)
        if vnet_id == 0 {
             WriteRequestErrorWithSubCode(w, http.StatusConflict, DEP_MISSING,
                   "VRF/VNET must be created prior to adding it to the VLAN interface" , []string{}, "")
             return
        }
    }

     /* Creation sequence:  1. Vlan, 2. Vlan Interface table, 3. Vlan Interface IP prefix table 4. Add local subnet route */
     /* Create 1 */
     vlan_pt := swsscommon.NewTable(db.swss_db, VLAN_TB)
     defer vlan_pt.Delete()
     vlan_pt.Set(vlan_name, map[string]string{
          "vlanid": vars["vlan_id"],
          "host_ifname": "Mon"+vlan_name,
     }, "SET", "")

    vlan_if_pt := swsscommon.NewTable(db.swss_db, VLAN_INTF_TB)
    defer vlan_if_pt.Delete()

    /* Create 2 */
    if attr.Vnet_id != "" {
        vnet_id_str = VNET_NAME_PREF + strconv.FormatUint(uint64(vnet_id), 10)
        vlan_if_pt.Set(vlan_name, map[string]string{
            "vnet_name": vnet_id_str,
            "proxy_arp": "enabled",
        }, "SET", "")
    }

    /* Create 3 */
    if attr.IPPrefix != "" {
        if attr.Vnet_id != "" {
            time.Sleep(time.Second)
        }
        vlan_if_pt.Set(generateDBTableKey(db.separator, vlan_name, attr.IPPrefix), map[string]string{"":""}, "SET", "")
        if attr.Vnet_id != "" {
             local_subnet_route_pt := swsscommon.NewProducerStateTable(app_db_ops.swss_db, LOCAL_ROUTE_TB)
             defer local_subnet_route_pt.Delete()
             // No error check for IPPrefix since it is already checked in unmarshal
             _, vlan_netw, _ := net.ParseCIDR(attr.IPPrefix)
             local_subnet_route_pt.Set(
                 generateDBTableKey(app_db_ops.separator, vnet_id_str, vlan_netw.String()),
                 map[string]string{"ifname": vlan_name},
             "SET", "")
        }
    }

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlansGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    var vnet_idMatch string
    var vnet_id string
    var VlansPerVnet []VlansPerVnetModel
    var VlansPerVnetReturn VlansPerVnetReturnModel
    if len(r.URL.Query()["vnet_id"]) <1 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnet_id"}, "No vnet_id specified")
        return
    }
    if len(r.URL.Query()["vnet_id"]) == 1 {
        vnet_id = r.URL.Query()["vnet_id"][0]
	var err error
	vnet_idMatch, _ ,err = get_and_validate_vnet_id(w,vnet_id)
	if err != nil {
	    return
        }
    } else if len(r.URL.Query()["vnet_id"]) > 1 {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"vnet_id"}, "May only specify one vnet_id")
        return
    }

    //Getting a map for all the entries that match VLAN_Interface
    vlan_map_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB,  "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    VlansPerVnetReturn.Vnet_id = vnet_id
    for k,_ := range vlan_map_kv{
        // adding 4 to the length for the maximum digit VLAN possible ex. 4095
        if len(k)<len(generateDBTableKey(db.separator,VLAN_INTF_TB,VLAN_NAME_PREF))+4+1{
              for _,value := range vlan_map_kv[k]{
                  if value == vnet_idMatch{
		     vlanId := k[len(generateDBTableKey(db.separator,VLAN_INTF_TB,VNET_NAME_PREF)):]
		     ip_prefix_raw,err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, VLAN_NAME_PREF+vlanId,"*"))
                     if err != nil {
                         WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
                         return
		     }
		     var ip_prefix string
		     // parsing through the key to get the ip_prefix for the vlanId that matches the given vnet_id
		     for prefix,_ := range ip_prefix_raw{
                         ip_prefix = prefix[len(generateDBTableKey(db.separator,VLAN_INTF_TB,VLAN_NAME_PREF+vlanId))+1:]

		     }
		     vlanInt,_ := strconv.Atoi(vlanId)
		     output := VlansPerVnetModel{
			 IPPrefix: ip_prefix,
			 VlanID: vlanInt,
	             }
                     VlansPerVnet = append(VlansPerVnet,output)
                  }
              }
        }

    }
    VlansPerVnetReturn.Attr = VlansPerVnet
    WriteRequestResponse(w, VlansPerVnetReturn, http.StatusOK)
}

func ConfigInterfaceVlansAllGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops

    var Vlans []VlansModel
    var VlansReturn VlansReturnModel

    //Getting a map for all the vlans in DB
    vlan_map_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_TB,  "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if vlan_map_kv == nil || len(vlan_map_kv) == 0 {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"Vlans"}, "")
        return
    }

    for _,v := range vlan_map_kv{
        vlan_name := VLAN_NAME_PREF + v["vlanid"]
        vlan_pref_kv, _ := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name, "*"))
        vlan_if_kv, _ := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_INTF_TB, vlan_name))

        var vnet_guid string
        if vlan_if_kv != nil {
            vnet_id := vlan_if_kv["vnet_name"]
            vmap, _ := GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vnet_id))
            vnet_guid = vmap["guid"]
        }

        var vlan_ip string
        if len(vlan_pref_kv) > 0 {
            for k,_ := range vlan_pref_kv {
                 ip_pref := k[(len(generateDBTableKey(db.separator,VLAN_INTF_TB, vlan_name)) + 1):]
                 ip, _, _ := net.ParseCIDR(ip_pref)
                 if IsValidIP(ip.String()) != true {
                     continue
                 }
                 vlan_ip = ip_pref
            }
        }

        vlanInt,_ := strconv.Atoi(v["vlanid"])
        output := VlansModel{
                      VlanID: vlanInt,
                      IPPrefix: vlan_ip,
                      Vnet_id: vnet_guid,
                  }
        Vlans = append(Vlans,output)
    }
    VlansReturn.Attr = Vlans
    WriteRequestResponse(w, VlansReturn, http.StatusOK)
}

func ConfigInterfaceVlansMembersAllGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops

    var Members = []VlanMembersModel{}
    var MembersReturn VlanMembersReturnModel
    var MembersAllReturn VlanMembersAllReturnModel

    //Getting a map for all the vlans in DB
    vlan_map_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_TB,  "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if vlan_map_kv == nil || len(vlan_map_kv) == 0 {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"Vlans"}, "")
        return
    }

    MembersAllReturn.Attr = make([]VlanMembersReturnModel, 0)
    for _,v := range vlan_map_kv{
        vlanInt,_ := strconv.Atoi(v["vlanid"])
        vlan_name := VLAN_NAME_PREF + v["vlanid"]
        // Getting all the key value pairs for VLAN_MEMBER|vlan_name*
        vlan_members_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name,"*"))
        if err != nil {
            WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
        }
        if len(vlan_members_kv) == 0 {
            continue
        }

        Members = nil
        for k,v := range vlan_members_kv{
            output := VlanMembersModel{
                If_name: k[len(generateDBTableKey(db.separator,VLAN_MEMB_TB,vlan_name))+1:],
                Tagging: v["tagging_mode"],
            }
            Members = append(Members,output)
        }
        MembersReturn.VlanID = vlanInt
        MembersReturn.Attr = Members
        MembersAllReturn.Attr = append(MembersAllReturn.Attr, MembersReturn)
    }    
    WriteRequestResponse(w, MembersAllReturn, http.StatusOK)
}


func ConfigInterfaceVlanMemberGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    vlan_id, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    vlan_member_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name, vars["if_name"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_member_kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"if_name"}, "")
        return
    }

    var attr VlanMemberModel
    attr.Tagging = vlan_member_kv["tagging_mode"]
    output := VlanMemberReturnModel{
        VlanID: vlan_id,
        If_name: vars["if_name"],
        Attr: attr,
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfaceVlanMemberDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    _, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]
    vlan_member_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name, vars["if_name"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_member_kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"if_name"}, "")
        return
    }

    vlan_member_pt := swsscommon.NewTable(db.swss_db, VLAN_MEMB_TB)
    defer vlan_member_pt.Delete()
    vlan_member_pt.Del(generateDBTableKey(db.separator, vlan_name, vars["if_name"]), "DEL", "")
    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlanMemberPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    var attr VlanMemberModel
    err := ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    /* Config validation and failure reporting */
    _, err = vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }

    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]
    vlan_members, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if attr.Tagging == "untagged" {
        for k, v := range vlan_members {
            if vars["if_name"] == strings.Split(k, db.separator)[2] && v["tagging_mode"] == "untagged" {
                WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
                  "Object already an untagged member of some vlan: " + vars["if_name"], []string{}, "")
                return
            }
        }
    }

    vlan_member_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name, vars["if_name"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vlan_member_kv != nil {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
                  "Object already a member of this vlan: " + vars["if_name"], []string{}, "")
        return
    }

    /* Config update */
    vlan_member_pt := swsscommon.NewTable(db.swss_db, VLAN_MEMB_TB)
    defer vlan_member_pt.Delete()

    vlan_member_pt.Set(generateDBTableKey(db.separator, vlan_name, vars["if_name"]),
                       map[string]string{"tagging_mode": attr.Tagging}, "SET", "")

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlanMembersGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)
    var Members = []VlanMembersModel{}
    var MembersReturn VlanMembersReturnModel

    vlan_id, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]
    // Getting all the key value pairs for VLAN_MEMBER|vlan_name*
    vlan_members_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_MEMB_TB, vlan_name,"*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
	return
    }
    if len(vlan_members_kv) == 0 {
	log.Printf("No members found for %v ", vlan_name)
        MembersReturn.VlanID = vlan_id
        MembersReturn.Attr = Members
        WriteRequestResponse(w, MembersReturn, http.StatusOK)
	return
    }
    for k,v := range vlan_members_kv{
        output := VlanMembersModel{
            If_name: k[len(generateDBTableKey(db.separator,VLAN_MEMB_TB,vlan_name))+1:],
            Tagging: v["tagging_mode"],
        }
        Members = append(Members,output)
    }
    MembersReturn.VlanID = vlan_id
    MembersReturn.Attr = Members
    WriteRequestResponse(w, MembersReturn, http.StatusOK)
}

func ConfigInterfaceVlanNeighborGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    if !IsValidIPBoth(vars["ip_addr"]) {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ip_addr"}, "")
        return
    }
    vlan_id, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    neigh_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_NEIGH_TB, vlan_name, vars["ip_addr"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if neigh_kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"ip_addr"}, "")
        return
    }

    output := VlanNeighborReturnModel{
        VlanID: vlan_id,
        Ip_addr: vars["ip_addr"],
    }

    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigInterfaceVlanNeighborDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    if !IsValidIPBoth(vars["ip_addr"]) {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ip_addr"}, "")
        return
    }

    _, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    neigh_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_NEIGH_TB, vlan_name, vars["ip_addr"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if neigh_kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"ip_addr"}, "")
        return
    }

    neigh_pt := swsscommon.NewTable(db.swss_db, VLAN_NEIGH_TB)
    defer neigh_pt.Delete()
    neigh_pt.Del(generateDBTableKey(db.separator, vlan_name, vars["ip_addr"]),"DEL", "")

    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlanNeighborPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)
    var family string

    if !IsValidIPBoth(vars["ip_addr"]) {
        WriteRequestError(w, http.StatusBadRequest, "Malformed arguments for API call", []string{"ip_addr"}, "")
        return
    }

    if IsValidIP(vars["ip_addr"]) {
        family = "IPv4"
    } else {
        family = "IPv6"
    }

    _, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    neigh_kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VLAN_NEIGH_TB, vlan_name, vars["ip_addr"]))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if neigh_kv != nil {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
              "Object already exists " + vars["ip_addr"], []string{}, "")
        return
    }

    /* Config update */
    neigh_pt := swsscommon.NewTable(db.swss_db, VLAN_NEIGH_TB)
    defer neigh_pt.Delete()

    neigh_pt.Set(generateDBTableKey(db.separator, vlan_name, vars["ip_addr"]),
                       map[string]string{"family": family}, "SET", "")
    w.WriteHeader(http.StatusNoContent)
}

func ConfigInterfaceVlanNeighborsGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)
    var Neighbors = []VlanNeighborsModel{}
    var NeighborsReturn VlanNeighborsReturnModel

    vlan_id, err := vlan_validator(w, vars["vlan_id"])
    if err != nil {
        // Error is already handled in this case
        return
    }
    vlan_name := VLAN_NAME_PREF + vars["vlan_id"]

    // Getting all the key value pairs for NEIGH|vlan_name*
    neighbors_kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VLAN_NEIGH_TB, vlan_name, "*"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if len(neighbors_kv) == 0 {
	log.Printf("No neighbors found for %v ", vlan_name)
        NeighborsReturn.VlanID = vlan_id
        NeighborsReturn.Attr = Neighbors
        WriteRequestResponse(w, NeighborsReturn, http.StatusOK)
	return
    }
    for k,_ := range neighbors_kv{
        output := VlanNeighborsModel{
            Ip_addr: k[len(generateDBTableKey(db.separator,VLAN_NEIGH_TB,vlan_name))+1:],
        }
        Neighbors = append(Neighbors,output)
    }
    NeighborsReturn.VlanID = vlan_id
    NeighborsReturn.Attr = Neighbors

    WriteRequestResponse(w, NeighborsReturn, http.StatusOK)
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
    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, "default_vxlan_tunnel"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestError(w, http.StatusNotFound, "Object not found", []string{"tunnel_type"}, "")
        return
    }

    pt := swsscommon.NewTable(db.swss_db, VXLAN_TUNNEL_TB)
    defer pt.Delete()
    pt.Del("default_vxlan_tunnel", "DEL", "")
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

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, "default_vxlan_tunnel"))
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

    var attr TunnelDecapModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    tunnel_name := "default_vxlan_tunnel"
    // Check if IP address is V4.
    if IsValidIP(attr.IPAddr) {
        tunnel_name = "default_vxlan_tunnel_v4"
    }

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, tunnel_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    //Check if tunnel already exist and the address family is same
    if kv != nil {
        if isV4orV6(kv["src_ip"]) == isV4orV6(attr.IPAddr) {
            WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
                   "Object already exists: Default Vxlan VTEP", []string{}, "")
            return
        }
    }

    pt := swsscommon.NewTable(db.swss_db, VXLAN_TUNNEL_TB)
    defer pt.Delete()

    pt.Set(tunnel_name, map[string]string{
        "src_ip": attr.IPAddr,
    }, "SET", "")

    CacheTunnelLpbkIps(attr.IPAddr, true)

    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelEncapVxlanVnidDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelEncapVxlanVnidGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusNoContent)
}

func ConfigTunnelEncapVxlanVnidPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    vars := mux.Vars(r)
    db := &conf_db_ops

    vnet_id_str, _, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
        return
    }

    vnet_dep, err := vnet_dependencies_exist(vnet_id_str)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
    if vnet_dep {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, DELETE_DEP,
              "Deleting object that has child dependency, child element must be deleted first", []string{}, "")
        return
    }

    pt := swsscommon.NewTable(db.swss_db, VNET_TB)
    defer pt.Delete()

    pt.Del(vnet_id_str, "DEL", "")
    CacheDeleteVnetGuidId(vars["vnet_name"])
    CacheDeletePrefixAdv(vnet_id_str)

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
    vars := mux.Vars(r)

    _, kv, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
        return
    }

    vnid, err := strconv.Atoi(kv["vni"])
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error, Non numeric vnid found in db", []string{}, "")
        return
    }

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    vnet_id_str := VNET_NAME_PREF + strconv.FormatUint(uint64(vnet_id), 10)
    kv, err = GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    output := VnetReturnModel {
        VnetName: vars["vnet_name"],
        Attr: VnetModel {
            Vnid: vnid,
            AdvPrefix: kv["advertise_prefix"],
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

    tunnel_name := "default_vxlan_tunnel_v4"
    kv_4, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, tunnel_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    tunnel_name = "default_vxlan_tunnel"
    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, tunnel_name))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil and kv_4 == nil {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, DEP_MISSING,
              "Default VxLAN VTEP must be created prior to creating VRF", []string{"tunnel"}, "")
        return
    }

    var v6_tunnel, v4_tunnel bool
    if kv_4 != nil {
        tunnel_name = "default_vxlan_tunnel_v4"
	v4_tunnel = true
    }
    if kv != nil {
        tunnel_name = "default_vxlan_tunnel"
	v6_tunnel = true
    }

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    if vnet_id != 0 {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
              "Object already exists: " + vars["vnet_name"], []string{}, "")
        return
    }

    guid := CacheGetVniId(uint32(attr.Vnid))
    if guid != "" {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
              "Object already exists: {\"vni\":\"" + strconv.Itoa(attr.Vnid) + "\", \"vnet_name\":\"" + guid +"\"}", []string{}, "")
        return
    }

    vnet_id = CacheGenAndSetVnetGuidId(vars["vnet_name"], uint32(attr.Vnid))
    vnet_id_str := VNET_NAME_PREF + strconv.FormatUint(uint64(vnet_id), 10)

    kv, err = GetKVs(db.db_num, generateDBTableKey(db.separator, VNET_TB, vnet_id_str))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error: GUID Cache and DB out of sync", []string{}, "")
        return
    }

    pt := swsscommon.NewTable(db.swss_db, VNET_TB)
    defer pt.Delete()
    
    log.Printf("debug: vnet_id_str: "+vnet_id_str)
    vnetParams := make(map[string]string)
    vnetParams["vxlan_tunnel"] = tunnel_name
    vnetParams["vni"] = strconv.Itoa(attr.Vnid)
    vnetParams["guid"] = vars["vnet_name"]
    if strings.Compare(vars["vnet_name"], "Vnet-default") == 0 {
        if v6_tunnel == false {
            WriteRequestError(w, http.StatusInternalServerError, "Vnet-default is for V6 Tunnels, please create Vnet-default-v4", []string{}, "")
            return
        }
        vnetParams["scope"] = "default"
        vnetParams["vxlan_tunnel"] = "default_vxlan_tunnel"
    } else if strings.Compare(vars["vnet_name"], "Vnet-default-v4") == 0 {
        if v4_tunnel == false {
            WriteRequestError(w, http.StatusInternalServerError, "V4 tunnel not created, please create V4 Vxlan Tunnel", []string{}, "")
            return
        }
        vnetParams["scope"] = "default"
        vnetParams["vxlan_tunnel"] = "default_vxlan_tunnel_v4"
    }
    if attr.AdvPrefix != "" {
        vnetParams["advertise_prefix"] = attr.AdvPrefix
        CacheSetPrefixAdv(vnet_id_str, attr.AdvPrefix)
    }
    if attr.OverlayDmac != "" {
        vnetParams["overlay_dmac"] = attr.OverlayDmac
    }
    pt.Set(vnet_id_str, vnetParams, "SET", "")        

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdRoutesDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &app_db_ops
    vars := mux.Vars(r)

    vnet_id_str, _, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
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

    routes, err := SwssGetVrouterRoutes(vnet_id_str, vnidMatch, "*")
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    var failed []RouteModel
    pt1 := swsscommon.NewProducerStateTable(db.swss_db, ROUTE_TUN_TB)
    defer pt1.Delete()
    pt2 := swsscommon.NewProducerStateTable(db.swss_db, LOCAL_ROUTE_TB)
    defer pt2.Delete()

    for _, r := range routes {
        table1 := generateDBTableKey(db.separator, vnet_id_str, r.IPPrefix)
        pt1.Del(table1, "DEL", "")
        table2 := generateDBTableKey(db.separator, vnet_id_str, r.IPPrefix)
        pt2.Del(table2, "DEL", "")
    }

    if len(failed) > 0 {
        output := RouteReturnModel {
            Failed:  failed,
        }
        WriteRequestResponse(w, output, http.StatusMultiStatus)
    } else {
        w.WriteHeader(http.StatusNoContent)
    }
}

func ConfigVrouterVrfIdRoutesGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    vars := mux.Vars(r)

    vnet_id_str, _, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
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

    routes, err := SwssGetVrouterRoutes(vnet_id_str, vnidMatch, ipprefix)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    WriteRequestResponse(w, routes, http.StatusOK)
}

func ConfigVrouterVrfIdRoutesPatch(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &app_db_ops
    vars := mux.Vars(r)
    var rt_tb_key string

    vnet_id_str, _, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
        return
    }

    var attr []RouteModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    var pt swsscommon.ProducerStateTable
    var rt_tb_name string
    defer pt.Delete()

    var failed []RouteModel

    tunnel_pt := swsscommon.NewProducerStateTable(db.swss_db, ROUTE_TUN_TB)
    defer tunnel_pt.Delete()
    local_pt := swsscommon.NewProducerStateTable(db.swss_db, LOCAL_ROUTE_TB)
    defer local_pt.Delete()

    for _, r := range attr {

        /*
        Reject incorrect CIDR address such as 10.20.30.4/24
        Accept only correct CIDR addresses such as 10.20.30.0/24 or 10.20.30.4/32 
        */

        ip, network, err := net.ParseCIDR(r.IPPrefix)
        if err != nil {
            r.Error_msg = "Incorrect IP Prefix"
            failed = append(failed, r)
            continue
        }

        if ip.String() != strings.Split(network.String(), "/")[0] {
            r.Error_msg = "Incorrect IP Prefix"
            failed = append(failed, r)
            continue
        }
        if adv_prefix, ok := CacheGetPrefixAdv(vnet_id_str); ok {
            if adv_prefix == "true" {
                prefix_len, _ := network.Mask.Size()
                if isV4orV6(ip.String()) == 4 {
                    if prefix_len < 18 {
                        r.Error_msg = "Prefix length lesser than 18 is not supported"
                        failed = append(failed, r)
                        continue                        
                    }
                } else {
                    if prefix_len < 64 {
                        r.Error_msg = "Prefix length lesser than 64 is not supported"
                        failed = append(failed, r)
                        continue                          
                    }
                }
            }
        }        
        if r.IfName == "" {
            pt = tunnel_pt
            rt_tb_name = ROUTE_TUN_TB
            if *RunApiAsLocalTestDocker {
                rt_tb_name = "_"+ROUTE_TUN_TB
            }            
        } else {
            pt = local_pt
            rt_tb_name = LOCAL_ROUTE_TB
            if *RunApiAsLocalTestDocker {
                rt_tb_name = "_"+LOCAL_ROUTE_TB
            }
        }

        bm_next_hop := isLocalTunnelNexthop(r.NextHop)
        if bm_next_hop {
            log.Printf("Skipping route %v as it is a /32 local subnet route", r)
            continue
        }

        rt_tb_key = generateDBTableKey(db.separator, rt_tb_name, vnet_id_str, r.IPPrefix)

        cur_route, err := GetKVs(db.db_num, rt_tb_key)/* generateDBTableKey(db.separator, ROUTE_TUN_TB, vnet_id_str, r.IPPrefix))*/
        if err != nil {
            r.Error_code = http.StatusInternalServerError
            r.Error_msg = "Internal service error"
            failed = append(failed, r)
        }
        if r.Cmd == "delete" {
            if cur_route == nil {
                r.Error_code = http.StatusNotFound
                    r.Error_msg = "Not found"
                    failed = append(failed, r)
            } else {
                    pt.Del(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), "DEL", "")
                }
        } else {
            if cur_route != nil {
                if r.IfName == "" {
                    if cur_route["endpoint"] != r.NextHop ||
                        cur_route["endpoint_monitor"] != r.NextHopMonitor ||
                        cur_route["primary"] != r.Primary ||
                        cur_route["adv_prefix"] != r.AdvPrefix ||
                        cur_route["mac_address"] != r.MACAddress ||
                        cur_route["vni"] != strconv.Itoa(r.Vnid) ||
                        cur_route["weight"] != r.Weight ||
                        cur_route["profile"] != r.Profile {
                            if r.Cmd == "add" {
                                /* Delete and re-add the route as it is not identical */
                                pt.Del(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), "DEL", "")
                            } else {
                                curr_endpoints := ExtractIPsFromString(cur_route["endpoint"])
                                curr_endpoint_monitors := ExtractIPsFromString(cur_route["endpoint_monitor"])
                                new_nexthops := ExtractIPsFromString(r.NextHop)
                                new_nexthop_monitors := ExtractIPsFromString(r.NextHopMonitor)
                                success := false;
                                if r.Cmd == "append" {
                                    /* Append to existing entry */
                                    for _, n_nxthop := range new_nexthops {
                                        if !IsPresentInSlice(curr_endpoints, n_nxthop) {
                                            curr_endpoints = append(curr_endpoints, n_nxthop)
                                            success = true
                                        }
                                    }
                                    for _, n_nxthop_mtr := range new_nexthop_monitors {
                                        if !IsPresentInSlice(curr_endpoint_monitors, n_nxthop_mtr) {
                                            curr_endpoint_monitors = append(curr_endpoint_monitors, n_nxthop_mtr)
                                            success = true
                                        }
                                    }
                                } else {
                                    /* Remove from existing entry */
                                    var ok bool;
                                    for _, n_nxthop := range new_nexthops {
                                        if ok, curr_endpoints = RemoveFromSlice(curr_endpoints, n_nxthop); !ok {
                                            r.Error_msg = n_nxthop+" not present to remove"
                                            failed = append(failed, r)
                                            break
                                        } else {
                                            success = true
                                        }
                                    }
                                    for _, n_nxthop_mtr := range new_nexthop_monitors {
                                        if ok, curr_endpoint_monitors = RemoveFromSlice(curr_endpoint_monitors, n_nxthop_mtr); !ok {
                                            r.Error_msg = n_nxthop_mtr+" not present to remove"
                                            failed = append(failed, r)
                                            break
                                        } else {
                                            success = true
                                        }
                                    }                                    
                                }                             
                                if len(curr_endpoint_monitors) > 0 && (len(curr_endpoint_monitors) != len(curr_endpoints)) {
                                    r.Error_msg = "there must be equal number of nexthop(s) and nexthop_monitor(s)"
                                    failed = append(failed, r)
                                    continue
                                }
                                if success == true {
                                    if len(curr_endpoints) == 0 {
                                        pt.Del(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), "DEL", "")
                                    } else {
                                        cur_route["endpoint"] = strings.Join(curr_endpoints, ",")
                                        if len(curr_endpoint_monitors) > 0 {
                                            cur_route["endpoint_monitor"] = strings.Join(curr_endpoint_monitors, ",")
                                        }                       
                                        pt.Set(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), cur_route, "SET", "")
                                    }
                                }
                                continue
                            }
                    } else {
                            /* Identical route */
                            continue
                    }
                } else {
                    if cur_route["ifname"] != r.IfName {
                        /* Delete and re-add the route as it is not identical */
                        pt.Del(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), "DEL", "")                        
                    } else {
                        /* Identical route */
                        continue
                    }
                }
            }
            /* Add new entry or append for the first time */
            if r.Cmd == "add" || r.Cmd == "append" {
                route_map := make(map[string]string)
                if r.IfName == "" {
                    route_map["endpoint"] = r.NextHop
                    if(r.MACAddress != "") {
                            route_map["mac_address"] = r.MACAddress
                    }
                    if(r.Vnid != 0) {
                            route_map["vni"] = strconv.Itoa(r.Vnid)
                    }
                } else {
                    route_map["ifname"] = r.IfName
                    if r.NextHop != "" {
                        route_map["nexthop"] = r.NextHop
                    }
                }
                if r.NextHopMonitor != "" {
                    route_map["endpoint_monitor"] = r.NextHopMonitor
                }
                if r.Primary != "" {
                    success := true
                    nexthops := ExtractIPsFromString(r.NextHop)
                    primaries := ExtractIPsFromString(r.Primary)
                    for _, primary := range primaries {
                        if !IsPresentInSlice(nexthops, primary) {
                            r.Error_msg = primary+" not present in nexthop list"
                            failed = append(failed, r)
                            success = false
                            break                            
                        }
                    }
                    if success == false {
                        continue
                    }
                    route_map["primary"] = r.Primary                    
                }
                if r.Weight != "" {
                    route_map["weight"] = r.Weight
                }
                if r.Profile != "" {
                    route_map["profile"] = r.Profile
                }
                if r.AdvPrefix != "" {
                    adv_ip, adv_network, err := net.ParseCIDR(r.AdvPrefix)
                    if err != nil {
                        r.Error_msg = "Incorrect Advertisement Prefix"
                        failed = append(failed, r)
                        continue
                    }

                    if adv_ip.String() != strings.Split(adv_network.String(), "/")[0] {
                        r.Error_msg = "Incorrect Advertisement Prefix"
                        failed = append(failed, r)
                        continue
                    }

		    prefix_len, _ := adv_network.Mask.Size()
                    if isV4orV6(adv_ip.String()) == 4 {
                        if prefix_len < 18 {
                            r.Error_msg = "Adv Prefix length lesser than 18 is not supported"
                            failed = append(failed, r)
                            continue
                        }
                    } else {
                        if prefix_len < 60 {
                            r.Error_msg = "Adv Prefix length lesser than 60 is not supported"
                            failed = append(failed, r)
                            continue
                        }
                    }
                    route_map["adv_prefix"] = r.AdvPrefix
                }
                if r.Monitoring != "" {
                    route_map["monitoring"] = r.Monitoring
                }
                pt.Set(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), route_map, "SET", "")
            } else {
                /* Remove from non existing entry */
                r.Error_msg = "Cannot remove from non-existing route. Please add the route first!"
                failed = append(failed, r)
                continue                
            }
		}
	}

    if len(failed) > 0 {
        output := RouteReturnModel {
            Failed:  failed,
        }
        WriteRequestResponse(w, output, http.StatusMultiStatus)
    } else {
        w.WriteHeader(http.StatusNoContent)
    }
}

func ConfigVrfVrfIdRoutesPatch(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var db *db_ops
    conf_db := &conf_db_ops
    app_db := &app_db_ops
    vars := mux.Vars(r)
    var rt_tb_key string
    vrf_id_str := vars["vrf_id"]

    if vrf_id_str != "default" {
        // Only default is supported
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    var attr []RouteModel

    err := ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    var pt swsscommon.Table
    conf_pt := swsscommon.NewTable(conf_db.swss_db, STATIC_ROUTE_TB)
    defer conf_pt.Delete()
    app_pt := swsscommon.NewTable(app_db.swss_db, STATIC_ROUTE_TB)
    defer app_pt.Delete()

    var failed []RouteModel

    for _, r := range attr {

        /*
        Reject incorrect CIDR address such as 10.20.30.4/24
        Accept only correct CIDR addresses such as 10.20.30.0/24 or 10.20.30.4/32 
        */

        ip, network, err := net.ParseCIDR(r.IPPrefix)
        if err != nil {
            r.Error_msg = "Incorrect IP Prefix"
            failed = append(failed, r)
            continue
        }

        if ip.String() != strings.Split(network.String(), "/")[0] {
            r.Error_msg = "Incorrect IP Prefix"
            failed = append(failed, r)
            continue
        }

        if r.Persistent == "true" {
            db = conf_db
            pt = conf_pt
        } else {
            db = app_db
            pt = app_pt
        }

        rt_tb_key = generateDBTableKey(db.separator, STATIC_ROUTE_TB, vrf_id_str, r.IPPrefix)

        cur_route, err := GetKVs(db.db_num, rt_tb_key)
        if err != nil {
            r.Error_code = http.StatusInternalServerError
            r.Error_msg = "Internal service error"
            failed = append(failed, r)
        }

        if r.Cmd == "delete" {
            if cur_route == nil {
                r.Error_code = http.StatusNotFound
                r.Error_msg = "Not found"
                failed = append(failed, r)
            } else {
                pt.Del(generateDBTableKey(db.separator, vrf_id_str, r.IPPrefix), "DEL", "")
            }
        } else {
            if cur_route != nil {
                if cur_route["nexthop"] != r.NextHop ||
                   cur_route["ifname"] != r.IfName {
                    /* Delete and re-add the route as it is not identical */
                    pt.Del(generateDBTableKey(db.separator,vrf_id_str, r.IPPrefix), "DEL", "")
                } else {
                    /* Identical route */
                    if r.Persistent == "true" {
                        continue
                    }
                }
            }
            route_map := make(map[string]string)
            if r.IfName == "" {
                route_map["nexthop"] = r.NextHop
            } else {
                route_map["ifname"] = r.IfName
                if r.NextHop != "" {
                    route_map["nexthop"] = r.NextHop
                }
            }
            if r.IfName == "null" {
                route_map["blackhole"] = "true"
            }
            if r.NextHopMonitor != "" {
                route_map["endpoint_monitor"] = r.NextHopMonitor
            }
            if r.Weight != "" {
                route_map["weight"] = r.Weight
            }
            if r.Profile != "" {
                route_map["profile"] = r.Profile
            }
            if r.Persistent == "false" {
                route_map["refresh"] = "true"
            }
            pt.Set(generateDBTableKey(db.separator,vrf_id_str, r.IPPrefix), route_map, "SET", "")
        }
    }

    if len(failed) > 0 {
        output := RouteReturnModel {
            Failed:  failed,
        }
        WriteRequestResponse(w, output, http.StatusMultiStatus)
    } else {
        w.WriteHeader(http.StatusNoContent)
    }
}

func ConfigVrfRouteExpiryPost(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &app_db_ops

    var attr RouteExpiryTimeModel
    err := ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    static_rt_t := swsscommon.NewTable(db.swss_db, STATIC_ROUTE_EXP_TB)
    defer static_rt_t.Delete()

    static_rt_t.Set("", map[string]string {
                        "time": strconv.Itoa(attr.Time),
        }, "SET", "")
    w.WriteHeader(http.StatusNoContent)    
}

func ConfigVrfRouteExpiryGet(w http.ResponseWriter, r *http.Request) {
    db := &app_db_ops

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, STATIC_ROUTE_EXP_TB))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if len(kv) == 0 {
        WriteRequestError(w, http.StatusBadRequest, "Object does not exist!", []string{}, "")
        return
    }

    time, _ := strconv.Atoi(kv["time"])
    output := RouteExpiryTimeModel {
        Time: time,
    }
    WriteRequestResponse(w, output, http.StatusOK)
}

func ConfigVrfVrfIdRoutesGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    vars := mux.Vars(r)
    vrf_id_str := vars["vrf_id"]

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

    app_db := &app_db_ops
    var pattern string

    pattern = generateDBTableKey(app_db.separator, STATIC_ROUTE_TB, vrf_id_str, ipprefix)
    kv1, err := GetKVsMulti(app_db.db_num, pattern)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    routes := []RouteModel{}
    for k, kvp := range kv1 {
        ipprefix, _ := ExtractIPPrefixFromKey(k, app_db.separator)
        routeModel := RouteModel{
            IPPrefix:    ipprefix,
            NextHop:     kvp["nexthop"],
        }

        if ifname, ok := kvp["ifname"]; ok {
            routeModel.IfName = ifname
        }

        if nexthop_monitor, ok := kvp["endpoint_monitor"]; ok {
            routeModel.NextHopMonitor = nexthop_monitor
        }
        
        if weight, ok := kvp["weight"]; ok {
            routeModel.Weight = weight
        }

        if profile, ok := kvp["profile"]; ok {
            routeModel.Profile = profile
        }

        routes = append(routes, routeModel)
    }

    conf_db := &conf_db_ops
    pattern = generateDBTableKey(conf_db.separator, STATIC_ROUTE_TB, vrf_id_str, ipprefix)
    kv2, err := GetKVsMulti(conf_db.db_num, pattern)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }
      
    for k, kvp := range kv2 {
        ipprefix, _ := ExtractIPPrefixFromKey(k, conf_db.separator)
        routeModel := RouteModel{
            IPPrefix:    ipprefix,
            NextHop:     kvp["nexthop"],
        }

        if ifname, ok := kvp["ifname"]; ok {
            routeModel.IfName = ifname
        }

        if nexthop_monitor, ok := kvp["endpoint_monitor"]; ok {
            routeModel.NextHopMonitor = nexthop_monitor
        }
        
        if weight, ok := kvp["weight"]; ok {
            routeModel.Weight = weight
        }

        if profile, ok := kvp["profile"]; ok {
            routeModel.Profile = profile
        }
        routeModel.Persistent = "true"
        routes = append(routes, routeModel)
    }

    WriteRequestResponse(w, routes, http.StatusOK)
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

// Required to run Unit Tests
func InMemConfigRestart(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    if *RunApiAsLocalTestDocker {
        genVnetGuidMap()
    }
    w.WriteHeader(http.StatusNoContent)
}

//Operations
func Ping(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var vnet_id_match string
    var out []byte
    var err error
    var attr PingRequestModel
    err_read_json := ReadJSONBody(w, r, &attr)
    if err_read_json != nil {
        // The error is already handled in this case
        return
    }
    if attr.VnetId != "" {
        vnet_id := attr.VnetId
	var err error
	vnet_id_match, _ ,err = get_and_validate_vnet_id(w,vnet_id)
	if err != nil {
	    // Error is handled in get_and_validate_vnet_id method
	    return
        }
    } else {
        log.Printf("vnet_id not provided \n")
    }

    var output PingReturnModel
    var count_param string
    if attr.Count != "" {
	count_param = "-c " + attr.Count
    } else  {
        log.Printf("count not provided , using default count 4 \n")
	count_param = "-c " + DEFAULT_PING_COUNT_STR
    }
    args := []string{attr.IpAddress, count_param}
    if vnet_id_match != "" {
	args = append(args, "-I", vnet_id_match)
    }
    out, err = exec.Command(PING_COMMAND_STR, args...).Output()
    if err != nil {
        log.Printf("Exec command Error is "+ err.Error())
    }

    op := string(out[:])
    output = parse_ping_output(op)

    WriteRequestResponse(w, output, http.StatusOK)
}
