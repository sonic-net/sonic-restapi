module go-server-server

go 1.14

require (
	github.com/comail/colog v0.0.0-20160416085026-fba8e7b1f46c
	github.com/go-redis/redis/v7 v7.3.0
	github.com/gorilla/mux v1.7.4
	github.com/satori/go.uuid v1.2.0
	github.com/vharitonsky/iniflags v0.0.0-20180513140207-a33cd0b5f3de
	swsscommon v0.0.0
)

replace swsscommon v0.0.0 => ../swsscommon
