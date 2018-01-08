.PHONY: all install build libcswsscommon

all: install build

install: build
	install -D $(GOPATH)/bin/go-server-server $(DESTDIR)/usr/sbin/go-server-server
	install -D $(GOPATH)/bin/mseethrifttest $(DESTDIR)/usr/sbin/thrift_mock_server
	install -D $(GOPATH)/bin/arpthrifttest $(DESTDIR)/usr/sbin/arp_mock_server

build: $(GOPATH)/bin/go-server-server $(GOPATH)/bin/mseethrifttest $(GOPATH)/bin/arpthrifttest

$(GOPATH)/bin/go-server-server: libcswsscommon $(GOPATH)/src/go-server-server/main.go
	cd $(GOPATH)/src/go-server-server && go get -v && go build -v

$(GOPATH)/bin/mseethrifttest:
	cd $(GOPATH)/src/mseethrifttest && go get -v && go build -v

$(GOPATH)/bin/arpthrifttest:
	cd $(GOPATH)/src/arpthrifttest && go get -v && go build -v

$(GOPATH)/src/go-server-server/main.go:
	mkdir -p            $(GOPATH)/src
	cp -r go-server-server $(GOPATH)/src/go-server-server
	cp -r arpthrift       $(GOPATH)/src/arpthrift
	cp -r mseethrift       $(GOPATH)/src/mseethrift
	cp -r swsscommon       $(GOPATH)/src/swsscommon
	mkdir -p $(GOPATH)/src/mseethrifttest
	cp mseethrifttest/server.go $(GOPATH)/src/mseethrifttest/server.go
	mkdir -p $(GOPATH)/src/arpthrifttest
	cp arpthrifttest/server.go $(GOPATH)/src/arpthrifttest/server.go

libcswsscommon:
	cd libcswsscommon && make install
