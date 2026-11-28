module go-server-server

go 1.26.0

require (
	github.com/comail/colog v0.0.0-20160416085026-fba8e7b1f46c
	github.com/gorilla/mux v1.8.1
	github.com/redis/go-redis/v9 v9.23.0
	github.com/satori/go.uuid v1.2.1-0.20180404165556-75cca531ea76
	github.com/vharitonsky/iniflags v0.0.0-20180513140207-a33cd0b5f3de
	swsscommon v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace swsscommon v0.0.0 => ../swsscommon
