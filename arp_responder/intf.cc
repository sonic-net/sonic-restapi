#include <sys/ioctl.h>
#include <arpa/inet.h>
#include <net/if.h>
#include <linux/if_packet.h>
#include <linux/if_ether.h>
#include <linux/filter.h>
#include <unistd.h>
#include <cstring>
#include <string>
#include <stdexcept>
#include "eintr.h"
#include "fmt.h"
#include "log.h"
#include "intf.h"


Interface::Interface(const std::string& iface_name) : name(iface_name),fd(-1)
{
    struct ifreq ifr;

    ifindex = ::if_nametoindex(name.c_str());
    if (ifindex == 0)
        throw std::runtime_error(s_fmt("Interface(%s)::constructor:if_nametoindex: error=(%d):%s", name.c_str(), errno, strerror(errno)));

    fd = ::socket(PF_INET, SOCK_STREAM, 0);
    if (fd == -1)
        throw std::runtime_error(s_fmt("Interface(%s)::constructor:socket: error=(%d):%s", name.c_str(), errno, strerror(errno)));

    (void) ::strncpy(ifr.ifr_name, iface_name.c_str(), sizeof(ifr.ifr_name) - 1);

    int res = ::ioctl(fd, SIOCGIFHWADDR, &ifr);
    if (res == -1)
        throw std::runtime_error(s_fmt("Interface(%s)::constructor:ioctl: error=(%d):%s", name.c_str(), errno, strerror(errno)));

    (void) ::memcpy(mac, ifr.ifr_hwaddr.sa_data, sizeof(mac));

    (void)HANDLE_EINTR(::close(fd));
}

void Interface::open(const struct sock_fprog* bpf)
{
    int err_code;
    struct sockaddr_ll addr;

    (void) ::memset(&addr, 0, sizeof(addr));
    addr.sll_family = AF_PACKET;
    addr.sll_ifindex = ifindex;

    fd = ::socket(AF_PACKET, SOCK_RAW, htons(ETH_P_ALL));
    if (fd == -1)
        throw std::runtime_error(s_fmt("Interface(%s)::open:socket: error=(%d):%s", name.c_str(), errno, strerror(errno)));

    int val = 1;
    err_code = ::setsockopt(fd, SOL_PACKET, PACKET_AUXDATA, &val, sizeof(val));
    if (err_code == -1)
    {
        close();
        throw std::runtime_error(s_fmt("Interface(%s)::open:setsockopt(auxdata): error=(%d):%s", name.c_str(), errno, strerror(errno)));
    }

    if (bpf != 0)
    {
        err_code = ::setsockopt(fd, SOL_SOCKET, SO_ATTACH_FILTER, bpf, sizeof(*bpf));
        if (err_code == -1)
        {
            close();
            throw std::runtime_error(s_fmt("Interface(%s)::open:setsockopt(bpf): error=(%d):%s", name.c_str(), errno, strerror(errno)));
        }
    }

    err_code = ::bind(fd, reinterpret_cast<struct sockaddr*>(&addr), sizeof(addr));
    if (err_code == -1)
    {
        close();
        throw std::runtime_error(s_fmt("Interface(%s)::open:bind: error=(%d):%s", name.c_str(), errno, strerror(errno)));
    }
}

void Interface::close()
{
    if (fd != -1)
    {
        (void)HANDLE_EINTR(::close(fd));
        fd = -1;
    }
}


ssize_t Interface::recv(uint8_t* buf, size_t len)
{
    struct iovec iov;
    struct msghdr msg;
    union {
        struct cmsghdr cmsg;
        char buf[CMSG_SPACE(sizeof(struct tpacket_auxdata))];
    } cmsg_buf;

    struct sockaddr_ll  from;

    msg.msg_name = &from;
    msg.msg_namelen = sizeof(from);
    msg.msg_iov = &iov;
    msg.msg_iovlen = 1;
    msg.msg_control = &cmsg_buf;
    msg.msg_controllen = sizeof(cmsg_buf);
    msg.msg_flags = 0;

    iov.iov_len = len;
    iov.iov_base = buf;

    ssize_t packet_len = HANDLE_EINTR(::recvmsg(fd, &msg, 0));
    if (packet_len == -1)
    {
        if (errno == ENETDOWN)
        {
            LOG_INFO("Interface %s is down", name.c_str());
            return -1;
        }
        else
            throw std::runtime_error(s_fmt("Interface(%s)::recv:recvmsg: error=(%d):%s", name.c_str(), errno, strerror(errno)));
    }

    for (struct cmsghdr *cmsg = CMSG_FIRSTHDR(&msg); cmsg; cmsg = CMSG_NXTHDR(&msg, cmsg))
    {
        struct tpacket_auxdata *aux;
        ssize_t r_len;
        struct vlan_tag *tag;

        if (cmsg->cmsg_len < CMSG_LEN(sizeof(struct tpacket_auxdata))
         || cmsg->cmsg_level != SOL_PACKET
         || cmsg->cmsg_type != PACKET_AUXDATA)
            continue;

        aux = (struct tpacket_auxdata *)CMSG_DATA(cmsg);
        if ((aux->tp_vlan_tci == 0) && !(aux->tp_status & TP_STATUS_VLAN_VALID))
            continue;

        r_len = packet_len > static_cast<ssize_t>(iov.iov_len) ? static_cast<ssize_t>(iov.iov_len) : packet_len;
        if (r_len < ETH_ALEN*2) break;

        (void) ::memmove(buf + ETH_ALEN*2 + sizeof(struct vlan_tag), buf + ETH_ALEN*2, packet_len - ETH_ALEN*2);

        tag = reinterpret_cast<struct vlan_tag *>(buf + ETH_ALEN*2);
        tag->vlan_tpid = htons(VLAN_TPID(aux, aux));
        tag->vlan_tci = htons(aux->tp_vlan_tci);
    }

    return packet_len;
}

ssize_t Interface::send(const uint8_t* buf, size_t len)
{
    struct sockaddr_ll to;
    struct sockaddr* to_ptr = reinterpret_cast<sockaddr*>(&to);
    to.sll_family = AF_PACKET;
    to.sll_ifindex = ifindex;
    to.sll_halen = 6;
    (void) ::memcpy(&to.sll_addr, buf, 6);

    ssize_t ret = HANDLE_EINTR(::sendto(fd, buf, len, 0, to_ptr, sizeof(to)));
    if (ret == -1)
    {
        if (errno == ENETDOWN)
        {
            LOG_INFO("Interface %s is down", name.c_str());
            return -1;
        }
        else
            throw std::runtime_error(s_fmt("Interface(%s)::send:sendto: error=(%d):%s", name.c_str(), errno, strerror(errno)));
    }

    return ret;
}


