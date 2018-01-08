#include <netinet/in.h>
#include <netinet/if_ether.h>
#include <cstring>
#include <string>
#include <sstream>
#include "intf.h"
#include "fmt.h"
#include "arp.h"


void Arp::populate_src_info(const Interface& iface)
{
    (void) ::memcpy(&eth_hdr->ether_shost, iface.get_mac(), ETH_ALEN);
    (void) ::memcpy(&arp_hdr->arp_sha, iface.get_mac(), ETH_ALEN);
}

void Arp::make_request(uint32_t tip, uint32_t sip)
{
    (void) ::memset(&eth_hdr->ether_dhost, 0xff, ETH_ALEN);

    arp_hdr->ea_hdr.ar_op = htons(ARPOP_REQUEST);

    uint32_t source_ip = htonl(sip);
    uint32_t requested_ip = htonl(tip);

    (void) ::memset(&arp_hdr->arp_tha, 0xff, ETH_ALEN);
    (void) ::memcpy(&arp_hdr->arp_spa, &source_ip, IP_ALEN);
    (void) ::memcpy(&arp_hdr->arp_tpa, &requested_ip, IP_ALEN);
}

void Arp::make_reply_from_request(const Interface& iface)
{
    (void) ::memcpy(&eth_hdr->ether_dhost, &eth_hdr->ether_shost, ETH_ALEN);

    arp_hdr->ea_hdr.ar_op = htons(ARPOP_REPLY);

    uint8_t tmp[IP_ALEN];
    (void) ::memcpy(tmp, &arp_hdr->arp_tpa, IP_ALEN);
    (void) ::memcpy(&arp_hdr->arp_tpa, &arp_hdr->arp_spa, IP_ALEN);
    (void) ::memcpy(&arp_hdr->arp_spa, tmp, IP_ALEN);

    (void) ::memcpy(&arp_hdr->arp_tha, &arp_hdr->arp_sha, ETH_ALEN);

    this->populate_src_info(iface);
}

std::string Arp::dump() const
{
    std::ostringstream stream;
    for(size_t i = 0; i < packet_len; i++)
        stream << s_fmt("%02x ", packet_ptr[i]);

    return stream.str();
}

MSEEArp::MSEEArp()
{
    this->init();
}

MSEEArp::MSEEArp(const Interface& iface, uint16_t stag, uint16_t ctag)
{
    this->init();

    eth_hdr->ether_type  = htons(ETHERTYPE_8021AD);
    stag_hdr->vlan_tci   = htons(stag);
    stag_hdr->ether_type = htons(ETHERTYPE_VLAN);
    ctag_hdr->vlan_tci   = htons(ctag);
    ctag_hdr->ether_type = htons(ETHERTYPE_ARP);
    
    arp_hdr->ea_hdr.ar_hrd = htons(ARPHRD_ETHER);
    arp_hdr->ea_hdr.ar_pro = htons(ETH_P_IP);
    arp_hdr->ea_hdr.ar_hln = ETH_ALEN;
    arp_hdr->ea_hdr.ar_pln = IP_ALEN;

    this->populate_src_info(iface);
}

void MSEEArp::init()
{
    packet_ptr = packet;
    packet_len = msee_arp_packet_len;
    bzero(packet, msee_arp_packet_len);
    eth_hdr = reinterpret_cast<struct ether_header*>(packet);
    stag_hdr = reinterpret_cast<struct vlan*>(&eth_hdr[1]);
    ctag_hdr = &stag_hdr[1];
    arp_hdr = reinterpret_cast<struct ether_arp*>(&ctag_hdr[1]);
}

RawArp::RawArp()
{
    this->init();
}

RawArp::RawArp(const Interface& iface)
{
    this->init();

    eth_hdr->ether_type = htons(ETHERTYPE_ARP);
    
    arp_hdr->ea_hdr.ar_hrd = htons(ARPHRD_ETHER);
    arp_hdr->ea_hdr.ar_pro = htons(ETH_P_IP);
    arp_hdr->ea_hdr.ar_hln = ETH_ALEN;
    arp_hdr->ea_hdr.ar_pln = IP_ALEN;

    this->populate_src_info(iface);
}

void RawArp::init()
{
    packet_ptr = packet;
    packet_len = bm_arp_packet_len;
    bzero(packet, bm_arp_packet_len);
    eth_hdr = reinterpret_cast<struct ether_header*>(packet);
    arp_hdr = reinterpret_cast<struct ether_arp*>(&eth_hdr[1]);
}
