.PHONY: all install build libcswsscommon

all: install build

install: build
	install -D $(GOPATH)/bin/go-server-server $(DESTDIR)/usr/sbin/go-server-server

build: $(GOPATH)/bin/go-server-server

$(GOPATH)/bin/go-server-server: libcswsscommon $(GOPATH)/src/go-server-server/main.go
	cd $(GOPATH)/src/go-server-server && go get -v && go build -v

$(GOPATH)/bin/mseethrifttest:
	cd $(GOPATH)/src/mseethrifttest && go get -v && go build -v

$(GOPATH)/bin/arpthrifttest:
	cd $(GOPATH)/src/arpthrifttest && go get -v && go build -v

$(GOPATH)/src/go-server-server/main.go:
	mkdir -p            $(GOPATH)/src
	cp -r go-server-server $(GOPATH)/src/go-server-server
	cp -r swsscommon       $(GOPATH)/src/swsscommon

libcswsscommon:
	cd libcswsscommon && make install
