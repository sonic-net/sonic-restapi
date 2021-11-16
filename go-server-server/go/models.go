package restapi

import (
    "encoding/json"
    "net"
    "strconv"
    "strings"
)

type HeartbeatReturnModel struct {
    ServerVersion   string `json:"server_version,omitempty"`
    ResetGUID       string `json:"reset_GUID,omitempty"`
    ResetTime       string `json:"reset_time,omitempty"`
    RoutesAvailable int    `json:"routes_available,omitempty"`
}

type ConfigResetStatusModel struct {
    ResetStatus      string `json:"reset_status,omitempty"`
}

type RouteModel struct {
    Cmd         string `json:"cmd,omitempty"`
    IPPrefix    string `json:"ip_prefix"`
    IfName      string `json:"ifname,omitempty"`
    NextHopType string `json:"nexthop_type,omitempty"`
    NextHop     string `json:"nexthop"`
    MACAddress  string `json:"mac_address,omitempty"`
    Vnid        int    `json:"vnid,omitempty"`
    Weight      string `json:"weight,omitempty"`
    Profile     string `json:"profile,omitempty"`
    Error_code  int    `json:"error_code,omitempty"`
    Error_msg   string `json:"error_msg,omitempty"`
}

type RouteReturnModel struct {
    Failed  []RouteModel `json:"failed,omitempty"`
}

type InterfaceModel struct {
    AdminState string `json:"admin-state"`
}

type InterfaceReturnModel struct {
    Port string         `json:"port"`
    Attr InterfaceModel `json:"attr"`
}

type VlanModel struct {
    Vnet_id  string  `json:"vnet_id,omitempty"`
    IPPrefix string  `json:"ip_prefix,omitempty"`
}

type VlanReturnModel struct {
    VlanID    int         `json:"vlan_id"`
    Attr      VlanModel   `json:"attr"`
}

type VlansModel struct {
    VlanID    int     `json:"vlan_id"`
    IPPrefix  string  `json:"ip_prefix,omitempty"`
    Vnet_id   string  `json:"vnet_id,omitempty"`
}

type VlansReturnModel struct {
    Attr      []VlansModel  `json:"attr"`
}

type VlanMemberModel struct {
    Tagging   string      `json:"tagging_mode"`
}

type VlanMemberReturnModel struct {
    VlanID    int              `json:"vlan_id"`
    If_name   string           `json:"if_name"`
    Attr      VlanMemberModel  `json:"attr"`
}

type VlanMembersModel struct {
    If_name   string           `json:"if_name"`
    Tagging   string           `json:"tagging_mode"`
}

type VlanMembersReturnModel struct {
    VlanID    int              `json:"vlan_id"`
    Attr      []VlanMembersModel  `json:"attr"`
}

type VlanNeighborReturnModel struct {
    VlanID    int              `json:"vlan_id"`
    Ip_addr   string           `json:"ip_addr"`
}

type VlanNeighborsModel struct {
    Ip_addr   string           `json:"ip_addr"`
}

type VlanNeighborsReturnModel struct {
    VlanID    int                  `json:"vlan_id"`
    Attr      []VlanNeighborsModel `json:"attr"`
}

type VlansPerVnetReturnModel struct {
    Vnet_id   string               `json:"vnet_id,omitempty"`
    Attr      []VlansPerVnetModel  `json:"attr"`
}

type VlansPerVnetModel struct {
    VlanID    int              `json:"vlan_id"`
    IPPrefix  string           `json:"ip_prefix,omitempty"`

}

type TunnelDecapModel struct {
    IPAddr string `json:"ip_addr"`
}

type TunnelDecapReturnModel struct {
    TunnelType string           `json:"tunnel_type"`
    Attr       TunnelDecapModel `json:"attr"`
}

type VnetModel struct {
    Vnid int `json:"vnid"`
}

type VnetReturnModel struct {
    VnetName string   `json:"vnet_id"`
    Attr VnetModel    `json:"attr"`
}

type PingRequestModel struct {
    IpAddress string   `json:"ip_addr"`
    VnetId string   `json:"vnet_id"`
    Count string   `json:"count"`
}

type PingReturnModel struct {
    PacketsTransmitted string   `json:"packets_transmitted"`
    PacketsReceived string   `json:"packets_received"`
    MinRTT string   `json:"min_rtt"`
    MaxRTT string   `json:"max_rtt"`
    AvgRTT string   `json:"avg_rtt"`
}

type ErrorInner struct {
    Code    int      `json:"code"`
    SubCode *int     `json:"sub-code,omitempty"`
    Message string   `json:"message"`
    Fields  []string `json:"fields,omitempty"`
    Details string   `json:"details,omitempty"`
}

type ErrorModel struct {
    Error ErrorInner `json:"error"`
}

type MissingValueError struct {
    Field string
}

type InvalidFormatError struct {
    Field   string
    Message string
}

func (e *MissingValueError) Error() string {
    return "JSON missing field: " + (*e).Field
}

func (e *InvalidFormatError) Error() string {
    return (*e).Message
}

func (m *RouteModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        Cmd         *string `json:"cmd"`
        IPPrefix    *string `json:"ip_prefix"`
        IfName      *string `json:"ifname"`
        NextHopType *string `json:"nexthop_type"`
        NextHop     *string `json:"nexthop"`
        MACAddress  *string `json:"mac_address"`
        Vnid        int     `json:"vnid"`
        Weight      *string `json:"weight"`
        Profile     *string `json:"profile"`
        Error       string  `json:"error"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }
    
    if required.Cmd == nil {
        err = &MissingValueError{"cmd"}
        return
    } else if required.IPPrefix == nil {
        err = &MissingValueError{"ip_prefix"}
        return
    } else if required.IfName == nil {
        if required.NextHop == nil {
            err = &MissingValueError{"nexthop"}
            return
        }
    } 

    if *required.Cmd != "add" && *required.Cmd != "delete" {
        err = &InvalidFormatError{Field: "cmd", Message: "Must be add/delete"}
        return
    }

    _, _, err = ParseIPBothPrefix(*required.IPPrefix)
    if err != nil {
        err = &InvalidFormatError{Field: "ip_prefix", Message: "Invalid IP prefix"}
        return
    }

    if required.NextHop != nil {
        if !strings.Contains(*required.NextHop, ",") && !IsValidIPBoth(*required.NextHop) {
            err = &InvalidFormatError{Field: "nexthop", Message: "Invalid IP address"}
            return
        }
        m.NextHop = *required.NextHop
    }

    if required.IfName == nil && required.MACAddress != nil {
        _, err = net.ParseMAC(*required.MACAddress)

        if err != nil {
            err = &InvalidFormatError{Field: "mac_address", Message: "Invalid MAC address"}
            return
        }
        m.MACAddress = *required.MACAddress
    }

    m.Cmd = *required.Cmd
    m.IPPrefix = *required.IPPrefix
    m.Vnid = required.Vnid
    if required.IfName != nil {
        m.IfName = *required.IfName
    }
    if required.Weight != nil {
        m.Weight = *required.Weight
    }
    if required.Profile != nil {
        m.Profile = *required.Profile
    }
    return
}

func (m *VlanMemberModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
         Tagging   string   `json:"tagging_mode,omitempty"`
   }{}
   err = json.Unmarshal(data, &required)
   if err != nil {
       return
   }

   if required.Tagging == "" {
       required.Tagging = "untagged"
   } else if required.Tagging != "untagged" && required.Tagging != "tagged" {
       err = &InvalidFormatError{Field: "tagging_mode", Message: "Invalid tagging_mode, must be tagged/untagged"}
       return
   }
   m.Tagging = required.Tagging
   return
}

func (m *VlanModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
         Vnet_id  string  `json:"vnet_id,omitempty"`
         IPPrefix string  `json:"ip_prefix,omitempty"`
   }{}
   err = json.Unmarshal(data, &required)
   if err != nil {
       return
   }
   m.Vnet_id = required.Vnet_id

   if required.IPPrefix != "" {
       _, _, err = ParseIPBothPrefix(required.IPPrefix)
       if err != nil {
             err = &InvalidFormatError{Field: "ip_prefix", Message: "Invalid IP prefix"}
             return
       }
       m.IPPrefix = required.IPPrefix
   } else {
       m.IPPrefix = ""
   }
   return
}

func (m *TunnelDecapModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        IPAddr *string `json:"ip_addr"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.IPAddr == nil {
        err = &MissingValueError{"ip_addr"}
        return
    }

    m.IPAddr = *required.IPAddr

    if !IsValidIPBoth(m.IPAddr) {
        err = &InvalidFormatError{Field: "ip_addr", Message: "Invalid IPv4 address"}
        return
    }

    return
}

func (m *VnetModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        Vnid *int `json:"vnid"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.Vnid == nil {
        err = &MissingValueError{"vnid"}
        return
    }

    if *required.Vnid >= 0x1000000 {
        err = &InvalidFormatError{Field: "vnid", Message: "vnid must be < 2^24"}
        return
    }

    m.Vnid = *required.Vnid

    return
}

func (m *PingRequestModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        IpAddress *string   `json:"ip_addr"`
        VnetId    string   `json:"vnet_id"`
        Count     string   `json:"count"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.IpAddress == nil {
        err = &MissingValueError{"ip_addr"}
        return
    }
    m.IpAddress = *required.IpAddress

    if !IsValidIPBoth(m.IpAddress) {
        err = &InvalidFormatError{Field: "ip_addr", Message: "Invalid IPv4 address"}
        return
    }
    if required.Count != "" {
        _,err_count := strconv.Atoi(required.Count)
	if err_count != nil {
            err = &InvalidFormatError{Field: "count", Message: "count should be an integer"}
	    return
	}
    }
    m.VnetId = required.VnetId
    m.Count = required.Count
    return
}
