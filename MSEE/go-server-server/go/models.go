package mseeserver

import (
    "encoding/json"
    "net"
)

type HeartbeatReturnModel struct {
    ServerVersion string `json:"server_version,omitempty"`
    ResetGUID    string `json:"reset_GUID,omitempty"`
    ResetTime     string `json:"reset_time,omitempty"`
}

type ServerSnapshotModel struct {
    DecapModel TunnelDecapModel   `json:"decap_model,omitempty"`
    VrfMap     map[int]VrfSnapshotModel `json:"vrf_map,omitempty"`
}

type VrfSnapshotModel struct {
    VrfInfo     VnetModel             `json:"vrf_info,omitempty"`
    VxlanMap    map[int]TunnelModel   `json:"vxlan_map,omitempty"`
    PortMap     map[string]PortModel  `json:"port_map,omitempty"`
    QinQPortMap map[string]QInQModel  `json:"qinqport_map,omitempty"`
    RoutesMap   map[string]RouteModel `json:"routes,omitempty"`
}

type RouteModel struct {
    IPPrefix    string `json:"ip_prefix"`
    NextHopType string `json:"nexthop_type"`
    NextHop     string `json:"nexthop"`
    MACAddress  string `json:"mac_address"`
    Vnid        int    `json:"vnid,omitempty"`
    SrcIP       string `json:"src_ip,omitempty"`
    Port        string `json:"port,omitempty"`
    Error       string `json:"error,omitempty"`
}

type RouteDeleteReturnModel struct {
    Removed []RouteModel `json:"removed,omitempty"`
    Failed  []RouteModel `json:"failed,omitempty"`
}

type RoutePutReturnModel struct {
    Added   []RouteModel `json:"added,omitempty"`
    Updated []RouteModel `json:"updated,omitempty"`
    Failed  []RouteModel `json:"failed,omitempty"`
}

type QInQModel struct {
    VrfID      int    `json:"vrf_id"`
    PeerIP     string `json:"peer_ip,omitempty"`
    ProxyArpIP string `json:"proxy_arp_ip,omitempty"`
    Subnet     string `json:"subnet,omitempty"`
}

type QInQReturnModel struct {
    Port string    `json:"port"`
    STag int       `json:"stag"`
    CTag int       `json:"ctag"`
    Attr QInQModel `json:"attr"`
}

type PortModel struct {
    VrfID      int      `json:"vrf_id"`
    Addr       string   `json:"addr,omitempty"`
    SpoofGuard []string `json:"spoof_guard,omitempty"`
    MACAddress string   `json:"mac_address,omitempty"`
}

type PortReturnModel struct {
    Port string    `json:"port"`
    Attr PortModel `json:"attr"`
}

type VlanModel struct {
    Vnet_id  string  `json:"vnet_id,omitempty"`
    IPPrefix string  `json:"ip_prefix,omitempty"`
}

type VlanReturnModel struct {
    VlanID    int         `json:"vlan_id"`
    Attr      VlanModel   `json:"attr"`
}

type VlanMemberModel struct {
    Tagging   string      `json:"tagging_mode"`
}

type VlanMemberReturnModel struct {
    VlanID    int              `json:"vlan_id"`
    If_name   string           `json:"if_name"`
    Attr      VlanMemberModel  `json:"attr"`
}

type TunnelDecapModel struct {
    IPAddr string `json:"ip_addr"`
}

type TunnelDecapReturnModel struct {
    TunnelType string           `json:"tunnel_type"`
    Attr       TunnelDecapModel `json:"attr"`
}

type VirtualRouterModel struct {
    VrfName    string   `json:"vrf_name"`
    DHCPRelays []string `json:"dhcp_relays,omitempty"`
}

type VirtualRouterReturnModel struct {
    VrfID int                `json:"vrf_id"`
    Attr  VirtualRouterModel `json:"attr"`
}

type InterfaceModel struct {
    AdminState string `json:"admin-state"`
}

type InterfaceReturnModel struct {
    Port string         `json:"port"`
    Attr InterfaceModel `json:"attr"`
}

type TunnelModel struct {
    VrfID int `json:"vrf_id"`
}

type TunnelReturnModel struct {
    Vnid int         `json:"vnid"`
    Attr TunnelModel `json:"attr"`
}

type VnetModel struct {
    Vnid int `json:"vnid"`
}

type VnetReturnModel struct {
    VnetName string   `json:"vnet_id"`
    Attr VnetModel    `json:"attr"`
}

type ErrorInner struct {
    Code    int      `json:"code"`
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
        IPPrefix    *string `json:"ip_prefix"`
        NextHopType *string `json:"nexthop_type"`
        NextHop     *string `json:"nexthop"`
        MACAddress  *string `json:"mac_address"`
        Vnid        int     `json:"vnid"`
        SrcIP       string  `json:"src_ip"`
        Port        string  `json:"port"`
        Error       string  `json:"error"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.IPPrefix == nil {
        err = &MissingValueError{"ip_prefix"}
        return
    } else if required.NextHopType == nil {
        err = &MissingValueError{"nexthop_type"}
        return
    } else if required.NextHop == nil {
        err = &MissingValueError{"nexthop"}
        return
    }

    if *required.NextHopType == "vxlan-tunnel" {
        if required.MACAddress == nil {
            err = &MissingValueError{"mac_address"}
            return
        }
        m.MACAddress = *required.MACAddress
    }

    m.IPPrefix = *required.IPPrefix
    m.NextHopType = *required.NextHopType
    m.NextHop = *required.NextHop
    m.Vnid = required.Vnid
    m.SrcIP = required.SrcIP
    m.Port = required.Port
    m.Error = required.Error

    _, _, err = ParseIPPrefix(m.IPPrefix)
    if err != nil {
        err = &InvalidFormatError{Field: "ip_prefix", Message: "Invalid IP prefix"}
        return
    }

    if m.NextHopType != "vxlan-tunnel" && m.NextHopType != "ip" {
        err = &InvalidFormatError{Field: "nexthop_type", Message: "nexthop_type may only be one of ip, vxlan-tunnel"}
        return
    }

    if !IsValidIP(m.NextHop) {
        err = &InvalidFormatError{Field: "nexthop", Message: "Invalid IP address"}
        return
    }

    if m.NextHopType == "vxlan-tunnel" {
        _, err = net.ParseMAC(m.MACAddress)
        if err != nil {
            err = &InvalidFormatError{Field: "mac_address", Message: "Invalid MAC address"}
            return
        }

        if m.Port == "" {
            m.Port = "azure"
        }

        if m.Port != "azure" && m.Port != "standard" {
            err = &InvalidFormatError{Field: "port", Message: "port may only be one of standard, azure"}
            return
        }
    }

    return
}

func (m *QInQModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        VrfID      *int   `json:"vrf_id"`
        PeerIP     string `json:"peer_ip"`
        ProxyArpIP string `json:"proxy_arp_ip"`
        Subnet     string `json:"subnet"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.VrfID == nil {
        err = &MissingValueError{"vrf_id"}
        return
    }

    m.VrfID = *required.VrfID
    m.PeerIP = required.PeerIP
    m.ProxyArpIP = required.ProxyArpIP
    m.Subnet = required.Subnet

    if !IsValidIP(m.PeerIP) {
        err = &InvalidFormatError{Field: "peer_ip", Message: "Invalid IPv4 address"}
        return
    }

    if !IsValidIP(m.ProxyArpIP) {
        err = &InvalidFormatError{Field: "proxy_arp_ip", Message: "Invalid IPv4 address"}
        return
    }

    _, _, err = ParseIPPrefix(m.Subnet)
    if err != nil {
        err = &InvalidFormatError{Field: "subnet", Message: "Invalid IPv4 prefix"}
        return
    }

    return
}

func (m *PortModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        VrfID      *int     `json:"vrf_id"`
        Addr       string   `json:"addr"`
        SpoofGuard []string `json:"spoof_guard"`
        MACAddress *string  `json:"mac_address"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    } else if required.VrfID == nil {
        err = &MissingValueError{"vrf_id"}
        return
    } else if required.MACAddress == nil {
        err = &MissingValueError{"mac_address"}
        return
    }

    m.VrfID = *required.VrfID
    m.Addr = required.Addr
    m.SpoofGuard = required.SpoofGuard
    m.MACAddress = *required.MACAddress

    _, _, err = ParseIPPrefix(m.Addr)

    if err != nil {
        err = &InvalidFormatError{Field: "addr", Message: "Invalid IPv4 prefix"}
        return
    }

    _, err = net.ParseMAC(m.MACAddress)

    if err != nil {
        err = &InvalidFormatError{Field: "mac_address", Message: "Invalid MAC address"}
        return
    }

    for _, addr := range m.SpoofGuard {
        _, _, err = ParseIPPrefix(addr)
        if err != nil {
            err = &InvalidFormatError{Field: "spoof_guard", Message: "Invalid IPv4 prefix"}
            return
        }
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

func (m *VirtualRouterModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        VrfName    *string  `json:"vrf_name"`
        DHCPRelays []string `json:"dhcp_relays"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    } else if required.VrfName == nil {
        err = &MissingValueError{"vrf_name"}
        return
    }

    m.VrfName = *required.VrfName
    m.DHCPRelays = required.DHCPRelays

    for _, ip := range m.DHCPRelays {
        if !IsValidIP(ip) {
            err = &InvalidFormatError{Field: "dhcp_relays", Message: "All entries in dhcp_relays must be a valid IPv4 address"}
            return
        }
    }

    return
}

func (m *InterfaceModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        AdminState *string `json:"admin-state"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    } else if required.AdminState == nil {
        err = &MissingValueError{"admin-state"}
        return
    }

    m.AdminState = *required.AdminState

    if m.AdminState != "up" && m.AdminState != "down" {
        err = &InvalidFormatError{Field: "admin-state", Message: "admin-state may only be one of up, down"}
        return
    }

    return
}

func (m *TunnelModel) UnmarshalJSON(data []byte) (err error) {
    required := struct {
        VrfID *int `json:"vrf_id"`
    }{}

    err = json.Unmarshal(data, &required)

    if err != nil {
        return
    }

    if required.VrfID == nil {
        err = &MissingValueError{"vrf_id"}
        return
    }

    m.VrfID = *required.VrfID

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
