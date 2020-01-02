#include <net/if.h>
#include <linux/if_ether.h>
#include <cstring>
#include <vector>
#include <set>
#include <unordered_map>
#include <map>
#include <thread>
#include <stdexcept>
#include <fstream>
#include <iostream>
#include "fmt.h"
#include "log.h"
#include "intf.h"
#include "poller.h"
#include "cmd.h"
#include "arpresponder_bm.h"


bool read_config(const std::string& filename, std::vector<std::string>& result)
{
    std::string line;

    std::ifstream infile(filename);
    if (!infile)
    {
        std::cerr << "Can't open file:" << filename << std::endl;
        return false;
    }

    while (std::getline(infile, line))
    {
        if (!line.empty())
            result.push_back(line);
    }

    return true;
}

int main()
{
    std::vector<std::string> interfaces;

    if (!read_config("/tmp/arpresponder.conf", interfaces))
        return -1;

    ARPResponder responder;
    for (auto interface: interfaces)
    {
        LOG_INFO("Adding interface: %s", interface.c_str());
        responder.add_interface(interface);
    }

    responder.run();

    return 0;
}
