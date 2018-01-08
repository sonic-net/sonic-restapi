typedef i16 vlan_tag_t
typedef i32 ip4_t
typedef binary mac_t

struct req_tuple_t
{
    1: string iface_name;
    2: vlan_tag_t stag;
    3: vlan_tag_t ctag;
}

struct req_tuples_t
{
    1: list<req_tuple_t> tuples;
    2: i32 index;
    3: ip4_t ip;
}

struct rep_tuple_t
{
    1: req_tuple_t request;
    2: i32 index;
    3: mac_t mac;     // presented when is_found is true
    4: bool is_found;
}

service arp_responder
{
    bool add_interface(1: string iface_name);
    bool del_interface(1: string iface_name);
    bool add_ip(1: string iface_name, 2: vlan_tag_t stag, 3: vlan_tag_t ctag, 4: ip4_t ip);
    bool del_ip(1: string iface_name, 2: vlan_tag_t stag, 3: vlan_tag_t ctag);
    list<rep_tuple_t> request_mac(1: list<req_tuples_t> requests);
}
