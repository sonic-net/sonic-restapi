#include <arpa/inet.h>
#include <cstdarg>
#include <cstring>
#include <string>
#include <stdexcept>
#include "fmt.h"


std::string s_fmt(const char *fmt, ...)
{
    char *result;
    va_list ap;

    ::va_start(ap, fmt);
    (void) ::vasprintf(&result, fmt, ap);
    ::va_end(ap);

    std::string str(result);
    ::free(result);

    return str;
}

std::string s_ip(uint32_t ip)
{
    char ip_str[16];
    uint32_t c_ip = htonl(ip);

    const char *result = ::inet_ntop(AF_INET, &c_ip, ip_str, sizeof(ip_str));
    if (result == NULL)
        throw std::runtime_error(s_fmt("s_ip::inet_ntop: error=(%d):%s", errno, strerror(errno)));

    return std::string(ip_str);
}

std::string s_mac(const uint8_t* mac)
{
    char mac_str[18];

    (void) snprintf(mac_str, sizeof(mac_str), "%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

    return std::string(mac_str);
}
