.INTERMEDIATE: arp_responder.thrift.i

all:    arpresponder_msee  arpresponder_msee_client arpresponder_bm

install: arpresponder_msee  arpresponder_msee_client arpresponder_bm
	/usr/bin/install -D arpresponder_msee_client $(DESTDIR)/usr/bin/arpresponder_msee_client
	/usr/bin/install -D arpresponder_msee $(DESTDIR)/usr/sbin/arpresponder_msee
	/usr/bin/install -D arpresponder_bm $(DESTDIR)/usr/sbin/arpresponder_bm

cppcheck:
	cppcheck --enable=all --inconclusive .

CFLAGS += --std=gnu++11 -ggdb -O0 -Wall -Wextra -Wlogical-op -Wold-style-cast -Wuseless-cast -Wshadow -Wformat=2 

arpresponder_msee_client: arpresponder_msee_client.o fmt.o arp_responder.o arp_responder_constants.o arp_responder_types.o
	g++ -o arpresponder_msee_client arpresponder_msee_client.o fmt.o arp_responder.o arp_responder_constants.o arp_responder_types.o -lthrift

arpresponder_msee: arpresponder_msee_main.o fmt.o intf.o poller.o arpresponder_msee.o cmd.o arp.o arp_responder.o arp_responder_constants.o arp_responder_types.o
	g++ -o arpresponder_msee arpresponder_msee_main.o fmt.o intf.o poller.o arpresponder_msee.o cmd.o arp.o arp_responder.o arp_responder_constants.o arp_responder_types.o -lpthread -lthrift

arpresponder_msee_client.o: arpresponder_msee_client.cc arp_responder.h
	g++ $(CFLAGS) -c $< -o $@

fmt.o: fmt.cc fmt.h log.h
	g++ $(CFLAGS) -c $< -o $@

intf.o: intf.cc intf.h log.h eintr.h
	g++ $(CFLAGS) -c $< -o $@

poller.o: poller.cc poller.h log.h eintr.h
	g++ $(CFLAGS) -c $< -o $@

cmd.o: cmd.cc cmd.h log.h eintr.h
	g++ $(CFLAGS) -c $< -o $@

arp.o: arp.cc arp.h log.h
	g++ $(CFLAGS) -c $< -o $@

arpresponder_msee_main.o: arpresponder_msee_main.cc arp_responder.h
	g++ $(CFLAGS) -c $< -o $@

arpresponder_msee.o: arpresponder_msee.cc arpresponder_msee.h cmd.h log.h arp.h
	g++ $(CFLAGS) -c $< -o $@

arp_responder.o: arp_responder.cpp arp_responder.h
	g++ $(CFLAGS) -c $< -o $@

arp_responder_constants.o: arp_responder_constants.cpp arp_responder_constants.h
	g++ $(CFLAGS) -c $< -o $@

arp_responder_types.o: arp_responder_types.cpp arp_responder_types.h
	g++ $(CFLAGS) -c $< -o $@

arp_responder.h arp_responder.cpp arp_responder_constants.h arp_responder_constants.cpp arp_responder_types.h arp_responder_types.cpp: arp_responder.thrift.i

arp_responder.thrift.i: arp_responder.thrift
	thrift --out . --gen cpp $<

arpresponder_bm: arpresponder_bm_main.o arpresponder_bm.o
	g++ -o arpresponder_bm arpresponder_bm_main.o arpresponder_bm.o poller.o intf.o arp.o fmt.o

arpresponder_bm.o: arpresponder_bm.cc arpresponder_bm.h fmt.h log.h intf.h poller.h
	g++ $(CFLAGS) -c $< -o $@

arpresponder_bm_main.o: arpresponder_bm_main.cc arpresponder_bm.h
	g++ $(CFLAGS) -c $< -o $@

clean:
	rm -f *.o arpresponder_msee  arpresponder_msee_client arpresponder_bm
	rm -f arp_responder.h arp_responder.cpp arp_responder_constants.h arp_responder_constants.cpp arp_responder_types.h arp_responder_types.cpp arp_responder_server.skeleton.cpp
