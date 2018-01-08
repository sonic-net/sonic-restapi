#ifndef __ARPRESPONDER_H
#define __ARPRESPONDER_H

const size_t MAX_NUM_OF_INTERFACES = 64;

typedef std::tuple<std::string, uint16_t, uint16_t> tag_key_t;
typedef std::tuple<uint16_t, uint16_t, uint32_t> waitlist_key_t;      // stag, ctag, ip
typedef std::tuple<int, uint32_t, std::set<int>> waitlist_value_t;    // index, requested_time, set<intf_fd>

enum mac_status_t
{
    MAC_NOT_READY,
    MAC_NOT_FOUND,
    MAC_FOUND
};

typedef std::tuple<mac_status_t, std::string, uint16_t, uint16_t, std::string> ready_value_t;
// is_ready, interface, stag, ctag, mac_address

class ARPResponder
{
public:
    explicit ARPResponder(int control_fd);
    ~ARPResponder();
    void run();
private:
    ARPResponder() {};
    ARPResponder(const ARPResponder&) {};
    void process(const int fd);
    void process_ctrl();
    void process_intf(const int fd);
    void add_interface(const struct cmd_request& request);
    void del_interface(const struct cmd_request& request);
    void request_mac_add(const struct cmd_request& request);
    void request_mac_complete(const struct cmd_request& request);
    void request_mac_do(const struct cmd_request& request);
    void request_mac_poll(const struct cmd_request& request);
    void timeout_requests();
    void add_ip(const struct cmd_request& request);
    void del_ip(const struct cmd_request& request);

    Poller* poller;
    Cmd*    cmd;
    std::unordered_map<int, Interface*> interfaces;
    std::unordered_map<std::string, int> fd_interfaces;

    std::map<tag_key_t, uint32_t> proxy_arp;

    std::vector<request_tuples_t> mac_requests;
    request_tuples_t mac_request;

    std::map<waitlist_key_t, waitlist_value_t> waitlist;
    std::map<int, ready_value_t> ready;

    int ctrl_fd;

    const int REPLY_TIMEOUT = 2;
};

#endif // __ARPRESPONDER_H
