# How to build the Go MSEE server

## Prerequesites
1) Install `libswsscommon-dev` and its matching runtime library, including the SDK Go adapter from [swss-common #1275](https://github.com/sonic-net/sonic-swss-common/pull/1275). A swss-common source checkout is not required.
2) libcswsscommon must be built, run make within `MSEE/libcswsscommon` after step 1.
2) Redis database on localhost:6379 - we use DB 0 (`APPL_DB`) and 4
3) Go version 1.8 (static binaries can be found at [https://golang.org/dl/](https://golang.org/dl/))
4) Thrift to build mock server - Ubuntu will need to build from source [https://g4greetz.wordpress.com/2016/10/29/how-to-install-apache-thrift-in-ubuntu/](https://g4greetz.wordpress.com/2016/10/29/how-to-install-apache-thrift-in-ubuntu/) - you may also need to add `/usr/local/lib` to `LD_LIBRARY_PATH`
5) Python (2 or 3) with python-redis for the tests
6) The following variables must be set
```
CGO_CFLAGS="-I(PATH_TO_NETWORKING_ACS_RESTAPI)/MSEE/libcswsscommon/include"

CGO_LDFLAGS="-L(PATH_TO_NETWORKING_ACS_RESTAPI)/MSEE/libcswsscommon -L(PATH_TO_LIBSWSSCOMMON_INSTALL)/lib"

LD_LIBRARY_PATH="(PATH_TO_NETWORKING_ACS_RESTAPI)/MSEE/libcswsscommon:(PATH_TO_LIBSWSSCOMMON_INSTALL)/lib"
```

## Go environment
The directories within the MSEE repo should be symlinked so that they follow the proper Go layout.

```
$GOPATH/src/go-server-server -> MSEE/go-server-server
$GOPATH/src/mseethrift       -> MSEE/mseethrift
$GOPATH/src/mseethrifttest   -> MSEE/mseethrifttest
$GOPATH/src/swsscommon       -> MSEE/swsscommon
```

Run `make build` from the repository root to prepare the SDK Go adapter in the local `swsscommon` module and build the `go-server-server` binary.

## Building the mock Thrift server
The Go server requires that it can connect to the Thrift server, we have a mock server that will allow every request to suceed, this will allow the Go server to run successfully.

To run the mock Thrift server run `go run server.go` within `$GOPATH/src/mseethrifttest`

## Command line arguments
The `go-server-server` binary supports several command line arguments.
```
-h                  - Lists command line arguments
-reset              - Resets the cache DB (DB 4)
-macaddr=<mac>      - Specifies switch MAC address (required)
-loaddr=<lo>        - Specified switch loopback address (required)
-thrift=<host:port> - Specifies the host and port of the Thrift server, defaults to localhost:9090
```

## Running tests
In order to run the tests the Redis DB and mock Thrift server must be running. Once they are both running, start the `go-server-server` binary.

In another terminal run `python -m unittest apitest` within the `MSEE/test` directory to run the tests. Some tests will be skipped or expected to fail, this is expected due to the incomplete API at this time. There should be no unexpected failures.

If there are unexpected failures ensure that the Redis DB and mock Thrift servers are running.
