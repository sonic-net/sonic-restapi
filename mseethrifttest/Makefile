#FIXME: Create better rules
#FIXME: Check and write all dependencies

RACE_OPTION := -race
ifeq ($(CONFIGURED_ARCH),armhf)
RACE_OPTION :=
endif

.INTERMEDIATE: msee.thrift.intermediate

all: client

client: client.go ../mseethrift/msee.go
	go build $(RACE_OPTION) client.go

msee/constants.go msee/msee.go msee/ttypes.go: msee.thrift.intermediate
	go get "git.apache.org/thrift.git/lib/go/thrift"

msee.thrift.intermediate: msee.thrift
	thrift -out . --gen go $<

clean:
	rm -fr msee
	rm -f client
