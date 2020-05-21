.PHONY: all install build libcswsscommon clean

GO := /usr/local/go/bin/go
export GOROOT=/usr/local/go
export GOPATH=$(HOME)/go
export GOBIN=$(GOPATH)/bin
export GO111MODULE=on

all: install build

install: build
	/usr/bin/install -D $(GOPATH)/bin/go-server-server debian/sonic-rest-api/usr/sbin/go-server-server

build: $(GOPATH)/bin/go-server-server

$(GOPATH)/bin/go-server-server: libcswsscommon $(GOPATH)/src/go-server-server/main.go
	cd $(GOPATH)/src/go-server-server && $(GO) get -v && $(GO) build -v

$(GOPATH)/src/go-server-server/main.go:
	mkdir -p               $(GOPATH)/src
	cp -r go-server-server $(GOPATH)/src/go-server-server
	cp -r swsscommon       $(GOPATH)/src/swsscommon

libcswsscommon:
	cd libcswsscommon && sudo make install

clean:
	rm -rf $(GOPATH)
