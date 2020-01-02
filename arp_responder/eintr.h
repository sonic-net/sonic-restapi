#ifndef __EINTR_H
#define __EINTR_H

#define HANDLE_EINTR(x) ({ \
    typeof(x) __eintr_result__; \
    do { \
    __eintr_result__ = (x); \
    } while (__eintr_result__ == -1 && errno == EINTR); \
    __eintr_result__; \
})

#endif // __EINTR_H
