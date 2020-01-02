#include <sys/epoll.h>
#include <cstring>
#include <vector>
#include <stdexcept>
#include "eintr.h"
#include "fmt.h"
#include "poller.h"


Poller::Poller()
{
    epoll_fd = ::epoll_create1(0);
    if (epoll_fd == -1)
        throw std::runtime_error(s_fmt("Poller::constructor:epoll_create1: error=(%d):%s", errno, strerror(errno)));

    cnt = 0;
}

void Poller::poll(std::vector<int>& ready)
{
    struct epoll_event events[MAX_EVENTS];

    int res = HANDLE_EINTR(::epoll_wait(epoll_fd, events, MAX_EVENTS, EPOLL_TIMEOUT));
    if (res == -1)
        throw std::runtime_error(s_fmt("Poller::poll:epoll_wait: error=(%d):%s", errno, strerror(errno)));

    for (int i = 0; i < res; i++)
        ready.push_back(events[i].data.fd);
}

void Poller::add_fd(const int fd)
{
    struct epoll_event ev;

    ev.events = EPOLLIN | EPOLLET;
    ev.data.fd = fd;
    int res = ::epoll_ctl(epoll_fd, EPOLL_CTL_ADD, fd, &ev);
    if (res == -1)
        throw std::runtime_error(s_fmt("Poller::add_fd:epoll_ctl: error=(%d):%s", errno, strerror(errno)));

    cnt++;
}

void Poller::del_fd(const int fd)
{
    int res = ::epoll_ctl(epoll_fd, EPOLL_CTL_DEL, fd, NULL);
    if (res == -1)
        throw std::runtime_error(s_fmt("Poller::del_fd:epoll_ctl: error=(%d):%s", errno, strerror(errno)));

    cnt--;
}
