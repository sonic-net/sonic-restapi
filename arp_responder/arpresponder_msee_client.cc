#include <iostream>

#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/transport/TSocket.h>
#include <thrift/transport/TTransportUtils.h>

#include "arp_responder.h"

#include "fmt.h"


using namespace apache::thrift;
using namespace apache::thrift::protocol;
using namespace apache::thrift::transport;

void prepare_tuples(std::vector<req_tuples_t>& tuples)
{
      req_tuples_t tuple;
      req_tuple_t r;

      r.iface_name = "eth0";
      r.stag = 10;
      r.ctag = 20;
      tuple.tuples.push_back(r);

      r.iface_name = "eth0";
      r.stag = 50;
      r.ctag = 100;
      tuple.tuples.push_back(r);

      tuple.index = 0;
      tuple.ip = 0x11223344;

      tuples.push_back(tuple);

      r.iface_name = "eth0";
      r.stag = 20;
      r.ctag = 30;
      tuple.tuples.push_back(r);

      r.iface_name = "eth0";
      r.stag = 60;
      r.ctag = 110;
      tuple.tuples.push_back(r);

      tuple.index = 1;
      tuple.ip = 0x55660000;

      tuples.push_back(tuple);
}

int main()
{
    boost::shared_ptr<TTransport> socket(new TSocket("localhost", 9091));
    boost::shared_ptr<TTransport> transport(new TBufferedTransport(socket));
    boost::shared_ptr<TProtocol> protocol(new TBinaryProtocol(transport));
    arp_responderClient* client = new arp_responderClient(protocol);

    try
    {
      bool res;
      transport->open();
      std::cout << "Starting client" << std::endl;

      res = client->add_interface("eth0");
      std::cout << "Added interface eth0. Result:" << res << std::endl;

      res = client->add_ip("eth0", 10, 20, 0x0a140001);
      std::cout << "Added ip 10.20.0.1 to eth0(10, 20). Result:" << res << std::endl;

      res = client->add_ip("eth0", 50, 100, 0x32640001);
      std::cout << "Added ip 50.100.0.1 to eth0(50, 100). Result:" << res << std::endl;

      std::vector<req_tuples_t> tuples;
      std::vector<rep_tuple_t> responses;
      prepare_tuples(tuples);
      client->request_mac(responses, tuples);

      for (auto r: responses)
      {
          std::cout << "Response index=" << r.index;
          if (r.is_found)
          {
              std::cout << " i=" << r.request.iface_name << " s=" << r.request.stag << " c=" << r.request.ctag;
              std::cout << " mac=" << s_mac(reinterpret_cast<const uint8_t*>(r.mac.data())) << std::endl;
          }
          else
          {
              std::cout << " mac wasn't found" << std::endl;
          }
      }

      res = client->del_ip("eth0", 50, 100);
      std::cout << "Deleted ip 50.100.0.1 to eth0(50, 100). Result:" << res << std::endl;

      res = client->del_ip("eth0", 10, 20);
      std::cout << "Deleted ip 10.20.0.1 to eth0(10, 20). Result:" << res << std::endl;

      res = client->del_interface("eth0");
      std::cout << "Deleted interface eth0. Result:" << res << std::endl;

      std::cout << "Stopping client" << std::endl;
    }
    catch (TException& tx)
    {
      std::cout << "ERROR: " << tx.what() << std::endl;
    }

    return 0;
}
