export GOROOT=/usr/local/go
export GOPATH=$HOME/go

mkdir -p               $GOPATH/src
cp -r go-server-server $GOPATH/src/go-server-server

go tool cover -html=coverage.cov -o=coverage.html