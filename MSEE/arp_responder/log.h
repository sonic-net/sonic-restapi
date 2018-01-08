#ifndef __LOG_H
#define __LOG_H

enum LOG_LEVEL {
  LOG_LEVEL_ERR,
  LOG_LEVEL_INFO,
  LOG_LEVEL_DEBUG    
};

#define DEFAULT_LOG_LEVEL LOG_LEVEL_INFO

#define LOG_ERR(format, ...) do { if (LOG_LEVEL_ERR <= DEFAULT_LOG_LEVEL) { LOG_TS(stderr, "ERR"); (void) ::fprintf(stderr, format, ##__VA_ARGS__); (void) ::fprintf(stderr, "\n"); ::fflush(stderr); }} while(0)
#define LOG_INFO(format, ...) do { if (LOG_LEVEL_INFO <= DEFAULT_LOG_LEVEL) { LOG_TS(stdout, "INFO"); (void) ::fprintf(stdout, format, ##__VA_ARGS__); (void) ::fprintf(stdout, "\n"); ::fflush(stdout); }} while(0)
#define LOG_DEBUG(format, ...) do { if (LOG_LEVEL_DEBUG <= DEFAULT_LOG_LEVEL) { LOG_TS(stdout, "DEBUG"); (void) ::fprintf(stdout, format, ##__VA_ARGS__); (void) ::fprintf(stdout, "\n"); ::fflush(stdout); }} while(0)
#define LOG_TS(where, level) do { \
                                  time_t raw; \
                                  struct tm *tm; \
                                  char ___buf[80]; \
                                  ::time(&raw); \
                                  tm = ::localtime(&raw); \
                                  ::strftime(___buf, sizeof(___buf), "%Y-%m-%d %I:%M:%S", tm); \
                                  ::fprintf(where, "%s: %s: ", ___buf, level); \
                                } while(0)

#endif // __LOG_H
