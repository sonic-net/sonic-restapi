#ifndef __ARPRESPONDER_BM_H
#define __ARPRESPONDER_BM_H

const size_t MAX_NUM_OF_INTERFACES = 64;

class ARPResponder
{
public:
    ARPResponder();
    ~ARPResponder();
    void add_interface(const std::string& iface_name);
    void run();
private:
    ARPResponder(const ARPResponder&) {};
    void process(const int fd);

    Poller* poller;

    std::unordered_map<int, Interface*> interfaces;
};

#endif // __ARPRESPONDER_BM_H
