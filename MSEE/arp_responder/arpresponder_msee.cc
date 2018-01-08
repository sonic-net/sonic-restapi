#include <netinet/in.h>
#include <netinet/if_ether.h>
#include <linux/if.h>
#include <linux/filter.h>
#include <set>
#include <map>
#include <unordered_map>
#include <tuple>
#include <algorithm>
#include "fmt.h"
#include "log.h"
#include "intf.h"
#include "poller.h"
#include "cmd.h"
#include "arp.h"
#include "arpresponder_msee.h"


ARPResponder::ARPResponder(int control_fd)
{
    poller = new Poller();
    cmd = new Cmd(control_fd);

    ctrl_fd = control_fd;
    poller->add_fd(ctrl_fd);

    mac_request = request_tuples_t();

    LOG_INFO("Starting arpresponder");
}

ARPResponder::~ARPResponder()
{
    LOG_INFO("Stopping arpresonder");

    for (auto intf_pair: interfaces)
        poller->del_fd(intf_pair.first);

    poller->del_fd(ctrl_fd);

    delete cmd;
    delete poller;
}

void ARPResponder::run()
{
    LOG_INFO("Starting main loop of arpresponder");

    std::vector<int> fds(MAX_NUM_OF_INTERFACES + 1);
    while(true)
    {
        fds.clear();
        poller->poll(fds);
        for(auto fd: fds)
            process(fd);
        timeout_requests();
    }
}

void ARPResponder::process(const int fd)
{
    if (fd == ctrl_fd)
        process_ctrl();
    else
        process_intf(fd);
}

void ARPResponder::process_ctrl()
{
    LOG_DEBUG("ARPResponder::received ctrl request");

    struct cmd_request request;

    cmd->recv_request(request);

    switch(request.code)
    {
        case ADD_INTERFACE:
            add_interface(request);
            break;
        case DEL_INTERFACE:
            del_interface(request);
            break;
        case ADD_IP:
            add_ip(request);
            break;
        case DEL_IP:
            del_ip(request);
            break;
        case REQUEST_MAC_ADD:
            request_mac_add(request);
            break;
        case REQUEST_MAC_COMPLETE:
            request_mac_complete(request);
            break;
        case REQUEST_MAC_DO:
            request_mac_do(request);
            break;
        case REQUEST_MAC_POLL:
            request_mac_poll(request);
            break;
        default:
            throw std::out_of_range(s_fmt("ARPResponder::process_ctrl: request code is invalid"));
    }
}

void ARPResponder::add_interface(const struct cmd_request& request)
{
    struct sock_filter arp_code_bpf[] = {
        { 0x20, 0, 0, 0xfffff030 }, // (000) ldh      vlan_avail
        { 0x15, 0, 5, 0x00000001 }, // (001) jeq      #0x1             jt 2    jf 7
        { 0x28, 0, 0, 0x0000000c }, // (002) ldh      [12]
        { 0x15, 0, 3, 0x00008100 }, // (003) jeq      #0x8100          jt 4    jf 7
        { 0x28, 0, 0, 0x00000010 }, // (004) ldh      [16]
        { 0x15, 0, 1, 0x00000806 }, // (005) jeq      #0x806           jt 6    jf 7
        { 0x6,  0, 0, 0x00040000 }, // (006) ret      #262144
        { 0x6,  0, 0, 0x00000000 }, // (007) ret      #0
    };
// generated manually based on tcpdump -ieth0 -d ether[12:2]=0x8100 and ether[16:2]=0x0806
#define ARRAY_SIZE(x) (sizeof(x) / sizeof((x)[0]))

    struct sock_fprog arp_bpf = {
        .len = ARRAY_SIZE(arp_code_bpf),
        .filter = arp_code_bpf
    };

    if (fd_interfaces.find(request.interface) == end(fd_interfaces))
    {
        LOG_INFO("Adding interface %s", request.interface);
        Interface* iface;
        try
        {
            iface = new Interface(request.interface);
        }
        catch(const std::runtime_error& e)
        {
            LOG_ERR("Invalid interface %s: %s", request.interface, e.what());
            cmd->resp_invalid_interface();
            return;
        }
        
        iface->open(&arp_bpf);
        int intf_fd = iface->get_fd();
        interfaces[intf_fd] = iface;
        fd_interfaces[request.interface] = intf_fd;
        poller->add_fd(intf_fd);
    }
    else
    {
        LOG_DEBUG("Interface %s already exists", request.interface);
    }
    cmd->resp_interface_added();
}

void ARPResponder::del_interface(const struct cmd_request& request)
{
    if (fd_interfaces.find(request.interface) != end(fd_interfaces))
    {
        LOG_INFO("Removing interface %s", request.interface);

        int intf_fd = fd_interfaces[request.interface];
        fd_interfaces.erase(request.interface);
        poller->del_fd(intf_fd);
        Interface* iface = interfaces[intf_fd];
        iface->close();
        delete iface;
    }
    else
    {
        LOG_DEBUG("Interface %s doesn't exist", request.interface);
    }

    cmd->resp_interface_deleted();
}

void ARPResponder::request_mac_add(const struct cmd_request& request)
{
    if (fd_interfaces.find(request.interface) == end(fd_interfaces))
    {
        cmd->resp_request_mac_added();
        return;
    }

    request_tuple_t t;

    t.iface_name = std::string(request.interface);
    t.stag = request.stag;
    t.ctag = request.ctag;

    mac_request.tuples.push_back(t);

    cmd->resp_request_mac_added();
}

void ARPResponder::request_mac_complete(const struct cmd_request& request)
{
    mac_request.index = request.index;
    mac_request.ip = request.ip;

    mac_requests.push_back(mac_request);
    mac_request.tuples.clear();

    cmd->resp_request_mac_completed();
}


void ARPResponder::request_mac_do(const struct cmd_request& request)
{
    (void) request;
    for (auto r: mac_requests)
        for (auto r1: r.tuples)
        {
            int intf_fd = fd_interfaces[r1.iface_name];
            Interface* iface = interfaces[intf_fd];

            auto proxy_key = tag_key_t(iface->get_name(), r1.stag, r1.ctag);

            if (proxy_arp.find(proxy_key) == end(proxy_arp))
            {
                LOG_ERR("Can't send ARP request on interface %s stag=%u ctag=%u. Interface ip doesn't exist",
                          iface->get_name().c_str(), r1.stag, r1.ctag);
                continue;
            }

            MSEEArp arp(*iface, r1.stag, r1.ctag);
            arp.make_request(r.ip, proxy_arp[proxy_key]);
            auto ret = iface->send(arp.get_packet(), arp.size());
            if (ret == -1) continue;

            LOG_INFO("Requesting mac address for ip %s. Interface=%s stag=%u ctag=%u",
                 s_ip(r.ip).c_str(), iface->get_name().c_str(), r1.stag, r1.ctag);

            waitlist_key_t key = std::make_tuple(r1.stag, r1.ctag, r.ip);
            if (waitlist.find(key) == end(waitlist))
                waitlist[key] = std::make_tuple(r.index, ::time(0), std::set<int>());
            std::get<2>(waitlist[key]).insert(intf_fd);
            ready[r.index] = std::make_tuple(MAC_NOT_READY, "", 0, 0, "");
        }

    mac_requests.clear();

    cmd->resp_request_mac_done();
}

void ARPResponder::request_mac_poll(const struct cmd_request& request)
{
    if (ready.find(request.index) == end(ready))
    {
        cmd->resp_request_mac_not_found(); // FIXME: invalid mac
        return;
    }

    if (std::get<0>(ready[request.index]) == MAC_NOT_READY)
    {
        cmd->resp_request_mac_not_ready();
        return;
    }

    if (std::get<0>(ready[request.index]) == MAC_NOT_FOUND)
        cmd->resp_request_mac_not_found();
    else
    {
        auto interface = std::get<1>(ready[request.index]);
        uint16_t stag = std::get<2>(ready[request.index]);
        uint16_t ctag = std::get<3>(ready[request.index]);
        const uint8_t* mac = reinterpret_cast<const uint8_t*>(std::get<4>(ready[request.index]).data());
        cmd->resp_request_mac_found(interface, stag, ctag, mac);
    }
}

void ARPResponder::timeout_requests()
{
    std::vector<waitlist_key_t> keys_for_removing;

    for (auto w: waitlist)
    {
        if (std::get<1>(w.second) + REPLY_TIMEOUT > time(0)) continue;

        auto stag = std::get<0>(w.first); 
        auto ctag = std::get<1>(w.first);
        auto requested_ip = std::get<2>(w.first);
        LOG_INFO("The request for ip %s (s=%d c=%d) was timed out",
                 s_ip(requested_ip).c_str(), stag, ctag);
        ready[std::get<0>(w.second)] = std::make_tuple(MAC_NOT_FOUND, "", 0, 0, "");
        keys_for_removing.push_back(w.first); 
    }

    for (auto k: keys_for_removing)
        waitlist.erase(k);
}

void ARPResponder::process_intf(const int fd)
{
    MSEEArp arp;
    Interface* iface = interfaces[fd];
    auto ret = iface->recv(arp.get_packet(), arp.size());
    if (ret == -1) return; 

    if (!arp.is_valid()) return;

    LOG_DEBUG("ARP packet dump %s", arp.dump().c_str());

    if (arp.get_type() == ARPOP_REQUEST)
    {
        auto key = tag_key_t(iface->get_name(), arp.get_stag(), arp.get_ctag());

        if (proxy_arp.find(key) == end(proxy_arp))
        {
            LOG_DEBUG("Got arp request on %s stag=%u ctag=%u with non-existent ip=%s",
                      iface->get_name().c_str(), arp.get_stag(), arp.get_ctag(), s_ip(arp.get_dst_ip()).c_str());
            return;
        }

        LOG_DEBUG("Responding on arp request on %s stag=%u ctag=%u ip=%s",
                  iface->get_name().c_str(), arp.get_stag(), arp.get_ctag(), s_ip(proxy_arp[key]).c_str());

        arp.make_reply_from_request(*iface);
        (void)iface->send(arp.get_packet(), arp.size());
    }

    if (arp.get_type() == ARPOP_REPLY)
    {
        waitlist_key_t key = std::make_tuple(arp.get_stag(), arp.get_ctag(), arp.get_src_ip());
        if (waitlist.find(key) == end(waitlist)) return;

        if (std::get<2>(waitlist[key]).count(iface->get_fd()) == 0) return;

        LOG_INFO("Got arp response on %s stag=%u ctag=%u for ip %s: %s",
                  iface->get_name().c_str(), arp.get_stag(), arp.get_ctag(), s_ip(arp.get_src_ip()).c_str(), s_mac(arp.get_src_mac()).c_str());
        int index = std::get<0>(waitlist[key]);
        auto mac = std::string(reinterpret_cast<const char*>(arp.get_src_mac()), ETH_ALEN);
        ready[index] = std::make_tuple(MAC_FOUND, iface->get_name(), arp.get_stag(), arp.get_ctag(), mac);
        waitlist.erase(key);
    }
}

void ARPResponder::add_ip(const struct cmd_request& request)
{
    if (request.stag == 0 || request.stag > 4095
     || request.ctag == 0 || request.ctag > 4095
     || fd_interfaces.find(request.interface) == end(fd_interfaces))
    {
        LOG_ERR("Invalid parameters for adding ip: i=%s stag=%u ctag=%u",
                request.interface, request.stag, request.ctag);
        cmd->resp_invalid_ip();
        return;
    }
    LOG_INFO("Adding ip %s stag:%u ctag:%u on %s",
              s_ip(request.ip).c_str(), request.stag, request.ctag, request.interface);

    auto key = tag_key_t(request.interface, request.stag, request.ctag);

    if (proxy_arp.find(key) != end(proxy_arp) && proxy_arp[key] != request.ip)
    {
        LOG_INFO("Overwrite ip address from %s to %s. i=%s s=%d c=%d",
                 s_ip(proxy_arp[key]).c_str(), s_ip(request.ip).c_str(),
                 request.interface, request.stag, request.ctag);
    }

    proxy_arp[key] = request.ip;
    cmd->resp_ip_added();
}

void ARPResponder::del_ip(const struct cmd_request& request)
{
    LOG_INFO("Removing ip stag:%u ctag:%u on %s",
              request.stag, request.ctag, request.interface);

    auto key = tag_key_t(request.interface, request.stag, request.ctag);
 
    if (proxy_arp.find(key) == end(proxy_arp))
    {
        cmd->resp_invalid_ip();
        return;
    }
    proxy_arp.erase(proxy_arp.find(key));
    cmd->resp_ip_deleted();
}

