typedef i32  msee_ip4_t
typedef i64  msee_mac_t
typedef i16  msee_vlan_t
typedef i64  msee_object_id_t
typedef i32  msee_vni_t
typedef i32  msee_vrf_id_t
typedef i16  msee_udp_port_t
typedef byte msee_port_t
typedef byte msee_prefix_len_t
typedef byte msee_port_count_t

struct msee_ip6_t
{
    1: i64 low;
    2: i64 high;
}

enum ip_type_t
{
    v4 = 4,
    v6 = 6
}

union msee_ip_t
{
    1: msee_ip4_t ip4;
    2: msee_ip6_t ip6;
}

struct msee_ip_address_t
{
    1: msee_ip_t ip;
    2: ip_type_t type;
}

struct msee_ip_prefix_t
{
    1: msee_ip_address_t ip;
    2: msee_prefix_len_t mask_length;
}

typedef string msee_group_t
typedef byte msee_queue_id_t
typedef string msee_counter_name
typedef string msee_statistics_name
typedef map<msee_group_t, map<msee_counter_name, i64>> counters_t;
typedef map<msee_group_t, map<msee_statistics_name, i64>> statistics_t;
typedef map<msee_queue_id_t, map<byte, double>> hist_t;

enum result_t
{
    OK,
    ERROR,
    ADDED,
    UPDATED,
    REMOVED,
    INVALID_PARAMETERS,
    NO_MEMORY,
    NOT_FOUND,
    ALREADY_EXISTS
}

service MSEE
{
    result_t init_dpdk_port(1: msee_port_count_t nb_customer_ports, 2: msee_mac_t mac_addr, 3: msee_ip4_t ipv4_loaddr, 4: msee_ip6_t ipv6_loaddr);

    // In MSEE scenario stag, ctag and port must all be set to correct values
    // In Baremetal, stag and ctag should be set to 0, port must be correct

    result_t add_port_to_vrf(1: msee_vrf_id_t vrf_id, 2: msee_port_t port, 3: msee_vlan_t outer_vlan, 4: msee_vlan_t inner_vlan);
    result_t delete_port_from_vrf(1: msee_port_t port, 2: msee_vlan_t outer_vlan, 3: msee_vlan_t inner_vlan);

    result_t map_vni_to_vrf(1: msee_vni_t vni, 2: msee_vrf_id_t vrf_id);
    result_t unmap_vni_to_vrf(1: msee_vni_t vni);

    result_t add_encap_route(1: msee_vrf_id_t vrf_id, 2: msee_ip_prefix_t dst_vm_ip_prefix, 3: msee_ip_address_t dst_host_ip, 4: msee_mac_t dst_mac_address, 5: msee_vni_t vni, 6: msee_udp_port_t port);
    result_t delete_encap_route(1: msee_vrf_id_t vrf_id, 2: msee_ip_prefix_t dst_vm_ip_prefix);

    result_t add_decap_route(1: msee_vrf_id_t vrf_id, 2: msee_ip_prefix_t dst_ip_prefix, 3: msee_mac_t mac, 4: msee_port_t port, 5: msee_vlan_t outer_vlan, 6: msee_vlan_t inner_vlan);
    result_t delete_decap_route(1: msee_vrf_id_t vrf_id, 2: msee_ip_prefix_t dst_ip_prefix);

    // group parameter could be "dpdk.total", "dpdk.switch_ports", "dpdk.nic", or "" to output everything
    counters_t get_counters(1: msee_group_t group);

    // group parameter could be "rings", "mempools", "fibs", or "" to output everything
    statistics_t get_statistics(1: msee_group_t group);
    hist_t get_hist();
}

// list of return values for every call which return result_t
//
// init_dpdk_port:       OK, INVALID_PARAMETERS
// add_port_to_vrf:      INVALID_PARAMETERS, ADDED, ALREADY_EXISTS, NO_MEMORY
// delete_port_from_vrf: INVALID_PARAMETERS, REMOVED, NOT_FOUND
// map_vni_to_vrf:       INVALID_PARAMETERS, ADDED, ALREADY_EXISTS, NO_MEMORY
// unmap_vni_to_vrf:     INVALID_PARAMETERS, REMOVED, NOT_FOUND
// add_encap_route:      INVALID_PARAMETERS, ADDED, UPDATED, NO_MEMORY
// delete_encap_route:   INVALID_PARAMETERS, REMOVED, NOT_FOUND
// add_decap_route:      INVALID_PARAMETERS, ADDED, UPDATED, NO_MEMORY
// delete_decap_route:   INVALID_PARAMETERS, REMOVED, NOT_FOUND