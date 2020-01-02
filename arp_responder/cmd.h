#ifndef __CMD_H
#define __CMD_H

enum cmd_request_code
{
    ADD_INTERFACE,
    DEL_INTERFACE,
    ADD_IP,
    DEL_IP,
    REQUEST_MAC_ADD,
    REQUEST_MAC_COMPLETE,
    REQUEST_MAC_DO,
    REQUEST_MAC_POLL
};

enum cmd_response_code
{
    INTERFACE_ADDED,
    INTERFACE_DELETED,
    INVALID_INTERFACE,
    IP_ADDED,
    IP_DELETED,
    INVALID_IP,
    REQUEST_MAC_ADDED,
    REQUEST_MAC_COMPLETED,
    REQUEST_MAC_DONE,
    REQUEST_MAC_FOUND,
    REQUEST_MAC_NOT_FOUND,
    REQUEST_MAC_NOT_READY
};

struct cmd_request
{
    cmd_request_code code;
    char interface[IFNAMSIZ];
    uint16_t stag;
    uint16_t ctag;
    uint32_t ip;
    int index;
};

struct cmd_response
{
    cmd_response_code code;
    char interface[IFNAMSIZ];
    uint16_t stag;
    uint16_t ctag;
    uint8_t mac[ETH_ALEN];
    int index;
};

struct request_tuple_t
{
    std::string iface_name;
    uint16_t stag;
    uint16_t ctag;
};

struct request_tuples_t
{
    std::vector<request_tuple_t> tuples;
    uint32_t ip;
    int index;
};

struct response_tuple_t
{
    request_tuple_t request;
    int index;
    uint8_t mac[ETH_ALEN];
    bool is_found;
};

const size_t request_len = sizeof(struct cmd_request);
const size_t response_len = sizeof(struct cmd_response);

class Cmd
{
public:
    explicit Cmd(int ctrl_fd) : fd(ctrl_fd) {}
    // send interface
    bool add_interface(const std::string& iface_name);
    bool del_interface(const std::string& iface_name);
    bool mod_interface(const std::string& iface_name, enum cmd_request_code code, enum cmd_response_code expected);
    bool add_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag, uint32_t ip);
    bool del_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag);
    bool mod_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag, uint32_t ip,
                enum cmd_request_code code, enum cmd_response_code expected);
    bool request_mac(std::vector<response_tuple_t>& responses, const std::vector<request_tuples_t>& requests);

    // receive interface
    void send_simple_response(enum cmd_response_code code);
    void resp_interface_added();
    void resp_interface_deleted();
    void resp_invalid_interface();
    void resp_ip_added();
    void resp_ip_deleted();
    void resp_invalid_ip();
    void resp_request_mac_added();
    void resp_request_mac_completed();
    void resp_request_mac_done();
    void resp_request_mac_found(std::string iface, uint16_t stag, uint16_t ctag, const uint8_t mac[ETH_ALEN]);

    void resp_request_mac_not_found();
    void resp_request_mac_not_ready();

    void recv_request(struct cmd_request& req);
private:
    Cmd() {};
    void request_mac_intf(const request_tuple_t& tuple, int index);
    void request_mac_complete(uint32_t ip);
    void request_mac_do();
    bool request_mac_poll(int index, std::string& iface, uint16_t& stag, uint16_t& ctag, uint8_t mac[ETH_ALEN], bool& is_found);
    void send_request(const struct cmd_request& req);
    void send_response(const struct cmd_response& resp);
    void recv_response(struct cmd_response& resp);
    void send(const char* buf, size_t size);
    void recv(char* buf, size_t size);

    int fd;
};

#endif // __CMD_H
