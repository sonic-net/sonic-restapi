package mseeserver

import (
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
    "swsscommon"
    "time"
    "github.com/gorilla/mux"
)

const RESRC_EXISTS int = 0
const DEP_MISSING int  = 1
const DELETE_DEP  int  = 2

func StateHeartbeatGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    output := HeartbeatReturnModel{
        ServerVersion: ServerAPIVersion,
        ResetGUID: ServerResetGuid,
        ResetTime: ServerResetTime,
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
            local_subnet_route_pt := swsscommon.NewTable(db.swss_db, CFG_LOCAL_ROUTE_TB)
            defer local_subnet_route_pt.Delete()
            local_subnet_route_pt.Del(generateDBTableKey(db.separator, vnet_id_str, vlan_netw.String()), "DEL", "")
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
     }, "SET", "")

    vlan_if_pt := swsscommon.NewTable(db.swss_db, VLAN_INTF_TB)
    defer vlan_if_pt.Delete()

    /* Create 2 */
    if attr.Vnet_id != "" {
        vnet_id_str = VNET_NAME_PREF + strconv.FormatUint(uint64(vnet_id), 10)
        vlan_if_pt.Set(vlan_name, map[string]string{
            "vnet_name": vnet_id_str,
        }, "SET", "")
    }

    /* Create 3 */
    if attr.IPPrefix != "" {
        if attr.Vnet_id != "" {
            time.Sleep(time.Second)
        }
        vlan_if_pt.Set(generateDBTableKey(db.separator, vlan_name, attr.IPPrefix), map[string]string{"":""}, "SET", "")
        if attr.Vnet_id != "" {
             local_subnet_route_pt := swsscommon.NewTable(db.swss_db, CFG_LOCAL_ROUTE_TB)
             defer local_subnet_route_pt.Delete()
             // No error check for IPPrefix since it is already checked in unmarshal
             _, vlan_netw, _ := net.ParseCIDR(attr.IPPrefix)
             local_subnet_route_pt.Set(
                 generateDBTableKey(db.separator, vnet_id_str, vlan_netw.String()),
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

    w.WriteHeader(http.StatusNoContent)
/* Comment out all the code below this point in this fn once Day 0 config for VTEP is complete */
    var attr TunnelDecapModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, "default_vxlan_tunnel"))
    if err != nil {
        /* WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "") */
        return
    }

    if kv != nil {
        /* WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS, 
               "Object already exists: Default Vxlan VTEP", []string{}, "") */
        return
    }

    pt := swsscommon.NewTable(db.swss_db, VXLAN_TUNNEL_TB)
    defer pt.Delete()

    pt.Set("default_vxlan_tunnel", map[string]string{
        "src_ip": attr.IPAddr,
    }, "SET", "")
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

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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

    kv, err := GetKVs(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, "default_vxlan_tunnel"))
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Internal service error", []string{}, "")
        return
    }

    if kv == nil {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, DEP_MISSING,
              "Default VxLAN VTEP must be created prior to creating VRF", []string{"tunnel"}, "")
        return
    }

    vnet_id := CacheGetVnetGuidId(vars["vnet_name"])
    if vnet_id != 0 {
        WriteRequestErrorWithSubCode(w, http.StatusConflict, RESRC_EXISTS,
              "Object already exists: " + vars["vnet_name"], []string{}, "")
        return
    }

    vnet_id = CacheGenAndSetVnetGuidId(vars["vnet_name"])
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

    pt.Set(vnet_id_str, map[string]string{
        "vxlan_tunnel": "default_vxlan_tunnel",
        "vni": strconv.Itoa(attr.Vnid),
        "guid": vars["vnet_name"],
    }, "SET", "")

    w.WriteHeader(http.StatusNoContent)
}

func ConfigVrouterVrfIdRoutesDelete(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db := &conf_db_ops
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
    pt := swsscommon.NewTable(db.swss_db, CFG_ROUTE_TUN_TB)
    defer pt.Delete()

    for _, r := range routes {
        table := generateDBTableKey(db.separator, vnet_id_str, r.IPPrefix)
        pt.Del(table, "DEL", "")
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
    db := &conf_db_ops
    vars := mux.Vars(r)
    var rt_tb_key string

    vnet_id_str, _, err := get_and_validate_vnet_id(w, vars["vnet_name"])
    if err != nil {
        // Error is already handled in this case
        return
    }

    vlanPrefixArr, err := generateVlanPrefixInVnet(vnet_id_str)
    if err != nil {
        WriteRequestError(w, http.StatusInternalServerError, "Error generating vlan prefix array", []string{}, "")
        return
    } else {
        log.Printf("info: gen vlanPrefixArr is %v", vlanPrefixArr)
    }

    var attr []RouteModel

    err = ReadJSONBody(w, r, &attr)
    if err != nil {
        // The error is already handled in this case
        return
    }

    pt := swsscommon.NewTable(db.swss_db, CFG_ROUTE_TUN_TB)
    defer pt.Delete()

    var failed []RouteModel

    for _, r := range attr {
        bm_next_hop, err := isBMNextHop(r.IPPrefix, vlanPrefixArr)
        if err != nil {
            r.Error_code = http.StatusInternalServerError
            r.Error_msg = "Internal service error"
            failed = append(failed, r)
            continue
        }
	     if bm_next_hop {
             log.Printf("Skipping route %v as it is a baremetal /32 route", r)
		       continue
		  }
        // TODO: Remove if else and correct getkvs in production code
        if *RunApiAsLocalTestDocker {
             rt_tb_key = generateDBTableKey(db.separator, "_"+ CFG_ROUTE_TUN_TB, vnet_id_str, r.IPPrefix)
        } else {
             rt_tb_key = generateDBTableKey(db.separator, CFG_ROUTE_TUN_TB, vnet_id_str, r.IPPrefix)
        }
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
                  if cur_route["endpoint"] != r.NextHop ||
                     cur_route["mac_address"] != r.MACAddress ||
                     cur_route["vni"] != strconv.Itoa(r.Vnid) {
                         /* Delete and re-add the route as it is not identical */
                         pt.Del(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), "DEL", "")
                  } else {
                         /* Identical route */
                         continue
                  }
             }
		       route_map := make(map[string]string)
				 route_map["endpoint"] = r.NextHop
				 if(r.MACAddress != "") {
				      route_map["mac_address"] = r.MACAddress
				 }
				 if(r.Vnid != 0) {
				      route_map["vni"] = strconv.Itoa(r.Vnid)
				 }
				 pt.Set(generateDBTableKey(db.separator,vnet_id_str, r.IPPrefix), route_map, "SET", "")

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
