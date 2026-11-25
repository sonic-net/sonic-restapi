.PHONY: all install build libcswsscommon clean

GO := /usr/local/go/bin/go
export GOROOT=/usr/local/go
export GOPATH=$(abspath .)/.gopath
export GOBIN=$(GOPATH)/bin
export GOMODCACHE=$(HOME)/go/pkg/mod
export GO111MODULE=on
export GOWORK=off
export GOTOOLCHAIN=local
export CGO_ENABLED=1

all: install build

install: build
	/usr/bin/install -D $(GOPATH)/bin/go-server-server debian/sonic-rest-api/usr/sbin/go-server-server
	/usr/bin/install -D $(GOPATH)/bin/go-server-server.test debian/sonic-rest-api/usr/sbin/go-server-server.test

build: $(GOPATH)/bin/go-server-server $(GOPATH)/bin/go-server-server.test

$(GOPATH)/bin/go-server-server: libcswsscommon $(GOPATH)/src/go-server-server/main.go
	cd $(GOPATH)/src/go-server-server && $(GO) get -v && $(GO) build $(RACE_OPTION) -v -o $(GOPATH)/bin/go-server-server

$(GOPATH)/bin/go-server-server.test: libcswsscommon $(GOPATH)/src/go-server-server/main.go
	cd $(GOPATH)/src/go-server-server && $(GO) get -v && $(GO) test $(RACE_OPTION) -c -covermode=atomic -coverpkg "go-server-server/go" -v -o $(GOPATH)/bin/go-server-server.test

$(GOPATH)/src/go-server-server/main.go:
	mkdir -p               $(GOPATH)/src
	cp -r go-server-server $(GOPATH)/src/go-server-server
	cp -r swsscommon       $(GOPATH)/src/swsscommon

libcswsscommon:
	make -C libcswsscommon
	sudo make -C libcswsscommon install

clean:
	rm -rf $(GOPATH)
	make -C libcswsscommon clean
