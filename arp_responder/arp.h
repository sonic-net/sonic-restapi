#ifndef __MAC_H
#define __MAC_H

#define ETHERTYPE_8021AD ETH_P_8021AD
#define IP_ALEN 4

struct vlan
{
    u_short vlan_tci;
    u_short ether_type;
};
// FIXME: can't find struct vlan

// QinQ ARP packet 
const int msee_arp_packet_len = sizeof(struct ether_header) + sizeof(struct vlan) + sizeof(struct vlan) + sizeof(struct ether_arp);

// RAW ARP packet
const int bm_arp_packet_len = sizeof(struct ether_header) + sizeof(struct ether_arp);

class Arp
{
public:
    void make_request(uint32_t tip, uint32_t sip);
    void make_reply_from_request(const Interface& iface);

    virtual bool is_valid() const=0;

    uint8_t get_type() const
    {
        return ntohs(arp_hdr->ea_hdr.ar_op);
    }

    uint8_t* get_packet() const
    {
        return packet_ptr;
    }

    size_t size() const
    {
        return packet_len;
    }

    uint8_t* get_src_mac() const
    {
        return arp_hdr->arp_sha;
    }

    uint8_t* get_dst_mac() const
    {
        return arp_hdr->arp_tha;
    }

    uint32_t get_src_ip() const
    {
        return ntohl(*reinterpret_cast<uint32_t*>(&arp_hdr->arp_spa));
    }

    uint32_t get_dst_ip() const
    {
        return ntohl(*reinterpret_cast<uint32_t*>(&arp_hdr->arp_tpa));
    }

    std::string dump() const;

protected:
    virtual void init()=0;
    void populate_src_info(const Interface& iface);

    struct ether_header* eth_hdr;
    struct ether_arp* arp_hdr;
    uint8_t* packet_ptr;
    size_t packet_len;
};


class MSEEArp : public Arp
{
public:
    MSEEArp();
    MSEEArp(const Interface& iface, uint16_t stag, uint16_t ctag);

    bool is_valid() const
    {
        if (eth_hdr->ether_type != htons(ETH_P_8021AD)) return false;
        if (stag_hdr->ether_type != htons(ETHERTYPE_VLAN)) return false;
        if (ctag_hdr->ether_type != htons(ETHERTYPE_ARP)) return false;

        return true; 
    }

    uint16_t get_stag() const
    {
        return ntohs(stag_hdr->vlan_tci);
    }

    uint16_t get_ctag() const
    {
        return ntohs(ctag_hdr->vlan_tci);
    }

private:
    void init();

    uint8_t packet[msee_arp_packet_len];
    struct vlan* stag_hdr;
    struct vlan* ctag_hdr;
};

class RawArp : public Arp
{
public:
    RawArp();
    explicit RawArp(const Interface& iface);

    bool is_valid() const
    {
        return eth_hdr->ether_type == htons(ETHERTYPE_ARP);
    }

private:
    void init();

    uint8_t packet[bm_arp_packet_len];
};

#endif // __MAC_H
