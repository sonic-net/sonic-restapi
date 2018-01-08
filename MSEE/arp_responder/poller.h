#ifndef __POLLER_H
#define __POLLER_H

const int MAX_EVENTS = 65;
const int EPOLL_TIMEOUT = 1000; // 1 second

class Poller
{
public:
    Poller();
    void poll(std::vector<int>& ready);
    void add_fd(const int fd);
    void del_fd(const int fd);
    int get_fd_counter() const
    {
        return cnt;
    }
private:
    int epoll_fd;
    int cnt;
};

#endif //__POLLER_H
