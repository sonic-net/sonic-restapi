#include <netinet/if_ether.h>
#include <linux/if.h>
#include <cstring>
#include <cassert>
#include <ctime>
#include <string>
#include <vector>
#include <unordered_set>
#include <stdexcept>
#include "eintr.h"
#include "fmt.h"
#include "cmd.h"


// send interface
bool Cmd::add_interface(const std::string& iface_name)
{
    return this->mod_interface(iface_name, ADD_INTERFACE, INTERFACE_ADDED);
}

bool Cmd::del_interface(const std::string& iface_name)
{
    return this->mod_interface(iface_name, DEL_INTERFACE, INTERFACE_DELETED);
}

bool Cmd::mod_interface(const std::string& iface_name, enum cmd_request_code code, enum cmd_response_code expected)
{
    struct cmd_request request;
    bzero(&request, request_len);

    request.code = code;
    (void) ::strncpy(request.interface, iface_name.c_str(), IFNAMSIZ);

    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);

    return response.code == expected;
}

bool Cmd::add_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag, uint32_t ip)
{
    return this->mod_ip(iface_name, stag, ctag, ip, ADD_IP, IP_ADDED);
}

bool Cmd::del_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag)
{
    return this->mod_ip(iface_name, stag, ctag, 0, DEL_IP, IP_DELETED);
}

bool Cmd::mod_ip(const std::string& iface_name, uint16_t stag, uint16_t ctag, uint32_t ip,
                 enum cmd_request_code code, enum cmd_response_code expected)
{
    struct cmd_request request;
    bzero(&request, request_len);

    request.code = code;
    (void) ::strncpy(request.interface, iface_name.c_str(), IFNAMSIZ);
    request.stag = stag;
    request.ctag = ctag;
    request.ip = ip;

    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);

    return response.code == expected;
}

bool Cmd::request_mac(std::vector<response_tuple_t>& responses, const std::vector<request_tuples_t>& requests)
{
    std::unordered_set<int> poll;
    for (auto request: requests)
    {
        for (auto tuple: request.tuples)
            request_mac_intf(tuple, request.index);

        request_mac_complete(request.ip);

        poll.insert(request.index);
    }

    request_mac_do();

    timespec pause_req = {
        .tv_sec = 0,
        .tv_nsec = 500000000,
    };
    timespec pause_rem;

    while (!poll.empty())
    {
        std::unordered_set<int> to_remove;

        for (auto index: poll)
        {
            response_tuple_t r;
            bool result = request_mac_poll(index, r.request.iface_name, r.request.stag, r.request.ctag, r.mac, r.is_found);
            if (result)
            {
                r.index = index;
                responses.push_back(r);
                to_remove.insert(index);
            }
        }

        for (auto index: to_remove)
            poll.erase(index);

        if (!poll.empty())
            (void) ::nanosleep(&pause_req, &pause_rem);
    }

    return true;
}

void Cmd::request_mac_intf(const request_tuple_t& tuple, int index)
{
    struct cmd_request request;
    ::bzero(&request, request_len);

    request.code = REQUEST_MAC_ADD;
    (void) ::strncpy(request.interface, tuple.iface_name.c_str(), IFNAMSIZ);
    request.stag = tuple.stag;
    request.ctag = tuple.ctag;
    request.index = index;

    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);

    if (response.code != REQUEST_MAC_ADDED)
        throw std::out_of_range(s_fmt("Cmd::request_mac_intf: response code is invalid"));    
}

void Cmd::request_mac_complete(uint32_t ip)
{
    struct cmd_request request;

    ::bzero(&request, request_len);
    request.code = REQUEST_MAC_COMPLETE;
    request.ip = ip;
    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);
    assert(response.code == REQUEST_MAC_COMPLETED);
}

void Cmd::request_mac_do()
{
    struct cmd_request request;

    ::bzero(&request, request_len);
    request.code = REQUEST_MAC_DO;
    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);
    assert(response.code == REQUEST_MAC_DONE);
}

bool Cmd::request_mac_poll(int index, std::string& iface, uint16_t& stag, uint16_t& ctag, uint8_t mac[ETH_ALEN], bool& is_found)
{
    struct cmd_request request;

    ::bzero(&request, request_len);
    request.code = REQUEST_MAC_POLL;
    request.index = index;
    this->send_request(request);

    struct cmd_response response;
    this->recv_response(response);
    if (response.code == REQUEST_MAC_NOT_READY) return false;

    if (response.code == REQUEST_MAC_FOUND)
    {
        iface = std::string(response.interface);
        stag = response.stag;
        ctag = response.ctag;
        (void) ::memcpy(mac, response.mac, ETH_ALEN);
        is_found = true;
    }
    else
    {
        assert(response.code == REQUEST_MAC_NOT_FOUND);
        ::bzero(mac, ETH_ALEN);
        is_found = false;
    }

    return true;
}

// receive interface
void Cmd::resp_interface_added()
{
    this->send_simple_response(INTERFACE_ADDED);
}

void Cmd::resp_interface_deleted()
{
    this->send_simple_response(INTERFACE_DELETED);
}

void Cmd::resp_invalid_interface()
{
    this->send_simple_response(INVALID_INTERFACE);
}

void Cmd::resp_ip_added()
{
    this->send_simple_response(IP_ADDED);
}

void Cmd::resp_ip_deleted()
{
    this->send_simple_response(IP_DELETED);
}

void Cmd::resp_invalid_ip()
{
    this->send_simple_response(INVALID_IP);
}

void Cmd::resp_request_mac_added()
{
    this->send_simple_response(REQUEST_MAC_ADDED);
}

void Cmd::resp_request_mac_completed()
{
    this->send_simple_response(REQUEST_MAC_COMPLETED);
}

void Cmd::resp_request_mac_done()
{
    this->send_simple_response(REQUEST_MAC_DONE);
}

void Cmd::resp_request_mac_found(std::string iface, uint16_t stag, uint16_t ctag, const uint8_t mac[ETH_ALEN])
{
    struct cmd_response resp;
    resp.code = REQUEST_MAC_FOUND;

    (void) ::strncpy(resp.interface, iface.c_str(), IFNAMSIZ);
    resp.stag = stag;
    resp.ctag = ctag;
    (void) ::memcpy(resp.mac, mac, ETH_ALEN);

    this->send_response(resp);
}

void Cmd::resp_request_mac_not_found()
{
    this->send_simple_response(REQUEST_MAC_NOT_FOUND);
}

void Cmd::resp_request_mac_not_ready()
{
    this->send_simple_response(REQUEST_MAC_NOT_READY);
}

void Cmd::send_simple_response(enum cmd_response_code code)
{
    struct cmd_response resp;
    resp.code = code;

    this->send_response(resp);
}

void Cmd::send_request(const struct cmd_request& req)
{
    this->send(reinterpret_cast<const char*>(&req), request_len);
}

void Cmd::recv_request(struct cmd_request& req)
{
    this->recv(reinterpret_cast<char*>(&req), request_len);
}

void Cmd::send_response(const struct cmd_response& resp)
{
    this->send(reinterpret_cast<const char*>(&resp), response_len);
}

void Cmd::recv_response(struct cmd_response& resp)
{
    this->recv(reinterpret_cast<char*>(&resp), response_len);
}

void Cmd::send(const char* buf, size_t size)
{
    ssize_t ret = HANDLE_EINTR(::send(fd, buf, size, 0));
    if (ret == -1)
        throw std::runtime_error(s_fmt("Cmd::send:send: error=(%d):%s", errno, strerror(errno)));
}

void Cmd::recv(char* buf, size_t size)
{
    ssize_t ret = HANDLE_EINTR(::recv(fd, buf, size, 0));
    if (ret == -1)
        throw std::runtime_error(s_fmt("Cmd::recv:send: error=(%d):%s", errno, strerror(errno)));
}

