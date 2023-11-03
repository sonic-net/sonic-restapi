#FIXME: Create better rules
#FIXME: Check and write all dependencies


.INTERMEDIATE: msee.thrift.intermediate

all: client

client: client.go ../mseethrift/msee.go
	go build -race client.go

msee/constants.go msee/msee.go msee/ttypes.go: msee.thrift.intermediate
	go get "git.apache.org/thrift.git/lib/go/thrift"

msee.thrift.intermediate: msee.thrift
	thrift -out . --gen go $<

clean:
	rm -fr msee
	rm -f client
