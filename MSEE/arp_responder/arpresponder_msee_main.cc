#include <net/if.h>
#include <linux/if_ether.h>
#include <cstring>
#include <vector>
#include <set>
#include <unordered_map>
#include <map>
#include <thread>
#include <stdexcept>
#include "fmt.h"
#include "log.h"
#include "intf.h"
#include "poller.h"
#include "cmd.h"
#include "arpresponder_msee.h"

#include "arp_responder.h"
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TSimpleServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>

using namespace ::apache::thrift;
using namespace ::apache::thrift::protocol;
using namespace ::apache::thrift::transport;
using namespace ::apache::thrift::server;

using boost::shared_ptr;

void arp_responder(int fd)
{
    ARPResponder responder(fd);
    responder.run();
}

class arp_responderHandler : virtual public arp_responderIf
{
public:
    arp_responderHandler()
    {
        int res = ::socketpair(AF_LOCAL, SOCK_DGRAM, 0, fd_);
        if (res == -1)
            throw std::runtime_error(s_fmt("Can't create socket pair: error=(%d):%s", errno, strerror(errno)));

        arp_ = std::thread(arp_responder, fd_[0]);
        cmd_ = new Cmd(fd_[1]);
    }

    ~arp_responderHandler()
    {
        (void) ::close(fd_[0]);
        (void) ::close(fd_[1]);
    }

    bool add_interface(const std::string& iface_name)
    {
        return cmd_->add_interface(iface_name);
    }

    bool del_interface(const std::string& iface_name)
    {
        return cmd_->del_interface(iface_name);
    }

    bool add_ip(const std::string& iface_name, const vlan_tag_t stag, const vlan_tag_t ctag, const ip4_t ip)
    {
        return cmd_->add_ip(iface_name, stag, ctag, ip);
    }

    bool del_ip(const std::string& iface_name, const vlan_tag_t stag, const vlan_tag_t ctag)
    {
        return cmd_->del_ip(iface_name, stag, ctag);
    }

    void request_mac(std::vector<rep_tuple_t>& _return, const std::vector<req_tuples_t>& requests)
    {
        std::vector<request_tuples_t> req;
        std::vector<response_tuple_t> resp;

        for (auto r: requests)
        {
            request_tuples_t rt;
            rt.index = r.index;
            rt.ip = r.ip;
            for (auto r1: r.tuples)
            {
                struct request_tuple_t rt1;
                rt1.iface_name = r1.iface_name;
                rt1.stag = r1.stag;
                rt1.ctag = r1.ctag;
                rt.tuples.push_back(rt1);
            }
            req.push_back(rt);
        }

        cmd_->request_mac(resp, req);

        for (auto r: resp)
        {
            rep_tuple_t rt;
            rt.index = r.index;
            rt.is_found = r.is_found;
            rt.mac = std::string(reinterpret_cast<char*>(r.mac), ETH_ALEN);
            rt.request.iface_name = r.request.iface_name;
            rt.request.stag = r.request.stag;
            rt.request.ctag = r.request.ctag;
            _return.push_back(rt);
        }
    }

private:
  int fd_[2];
  std::thread arp_;
  Cmd* cmd_;
};

int main()
{
  int port = 9091;
  shared_ptr<arp_responderHandler> handler(new arp_responderHandler());
  shared_ptr<TProcessor> processor(new arp_responderProcessor(handler));
  shared_ptr<TServerTransport> serverTransport(new TServerSocket(port));
  shared_ptr<TTransportFactory> transportFactory(new TBufferedTransportFactory());
  shared_ptr<TProtocolFactory> protocolFactory(new TBinaryProtocolFactory());

  TSimpleServer server(processor, serverTransport, transportFactory, protocolFactory);
  server.serve();
  return 0;
}

