#ifndef __INTF_H
#define __INTF_H

struct vlan_tag
{
    u_short vlan_tpid;
    u_short vlan_tci;
};

#define VLAN_TPID(hdr, hv) (((hv)->tp_vlan_tpid || ((hdr)->tp_status & TP_STATUS_VLAN_TPID_VALID)) ? (hv)->tp_vlan_tpid : ETH_P_8021Q)

class Interface
{
public:
    explicit Interface(const std::string& iface_name);
    void    open(const struct sock_fprog* bpf = 0);
    void    close();
    ssize_t recv(uint8_t* buf, size_t len);
    ssize_t send(const uint8_t* buf, size_t len);

    int get_fd() const
    {
        return fd;
    }

    const uint8_t* get_mac() const
    {
        return mac;
    }

    const std::string& get_name() const
    {
        return name;
    }

private:
    Interface() {};

    std::string name;
    int fd;
    unsigned int ifindex;
    uint8_t mac[6];
};


#endif // __INTF_H
