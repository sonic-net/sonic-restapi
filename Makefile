.PHONY: all install build libcswsscommon clean FORCE

GO := /usr/local/go/bin/go
export GOROOT=/usr/local/go
export GOPATH=$(HOME)/go
export GOBIN=$(GOPATH)/bin
export GO111MODULE=on
export GOFLAGS=-modcacherw
REDISCONN_GO := $(GOPATH)/src/swsscommon/redisconn/dial.go
RACE_OPTION := -race
ifeq ($(CONFIGURED_ARCH),armhf)
RACE_OPTION :=
endif

all: install build

install: build
	/usr/bin/install -D $(GOPATH)/bin/go-server-server debian/sonic-rest-api/usr/sbin/go-server-server
	/usr/bin/install -D $(GOPATH)/bin/go-server-server.test debian/sonic-rest-api/usr/sbin/go-server-server.test

build: $(GOPATH)/bin/go-server-server $(GOPATH)/bin/go-server-server.test

$(GOPATH)/bin/go-server-server: libcswsscommon $(REDISCONN_GO)
	cd $(GOPATH)/src/go-server-server && $(GO) get -v && $(GO) build $(RACE_OPTION) -v -o $(GOPATH)/bin/go-server-server

$(GOPATH)/bin/go-server-server.test: libcswsscommon $(REDISCONN_GO)
	cd $(GOPATH)/src/go-server-server && $(GO) get -v && $(GO) test $(RACE_OPTION) -c -covermode=atomic -coverpkg "go-server-server/go" -v -o $(GOPATH)/bin/go-server-server.test

$(GOPATH)/src/go-server-server/main.go:
	mkdir -p               $(GOPATH)/src
	cp -r go-server-server $(GOPATH)/src/go-server-server
	cp -r swsscommon       $(GOPATH)/src/swsscommon

$(REDISCONN_GO): /usr/share/swss/go/redisconn/dial.go $(GOPATH)/src/go-server-server/main.go FORCE
	cmp -s $< $@ || /usr/bin/install -D -m 0644 $< $@

FORCE:

libcswsscommon:
	make -C libcswsscommon
	sudo make -C libcswsscommon install

clean:
	rm -rf $(GOPATH)
	make -C libcswsscommon clean
