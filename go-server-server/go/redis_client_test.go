package restapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

const authRoot = "/run/redis-auth"

func requireRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 || os.Getenv("SWSS_AUTH_TEST_CONTAINER") != "1" {
		t.Skip("requires root in an isolated container")
	}
	if _, err := os.Lstat(authRoot); !os.IsNotExist(err) {
		t.Fatal("refuse to replace existing runtime state")
	}
	t.Cleanup(func() { os.RemoveAll(authRoot) })
}

func fixture(t *testing.T) (string, func()) {
	t.Helper()
	requireRoot(t)
	directory, err := os.MkdirTemp("/tmp", "rest-redis-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(directory) })
	if err := os.Chmod(directory, 0755); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(directory, "redis.sock")
	if err := os.MkdirAll(authRoot+"/local-system", 0755); err != nil {
		t.Fatal(err)
	}
	policy := []map[string]interface{}{{"username": "sonic-trusted-writer", "unix_socket": socket,
		"tcp_host": "127.0.0.1", "tcp_port": 0}}
	data, _ := json.Marshal(policy)
	if err := os.WriteFile(authRoot+"/local.json", data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authRoot+"/local-system/writer.secret", []byte("rest-local-test-password"), 0400); err != nil {
		t.Fatal(err)
	}
	acl := filepath.Join(directory, "users.acl")
	if err := os.WriteFile(acl, []byte("user default on nopass ~* &* -@all +@read +@transaction +select +ping\nuser sonic-trusted-writer on >rest-local-test-password ~* &* +@all\n"), 0600); err != nil {
		t.Fatal(err)
	}
	process := exec.Command("redis-server", "--port", "0", "--unixsocket", socket, "--unixsocketperm", "666",
		"--save", "", "--appendonly", "no", "--aclfile", acl)
	var output bytes.Buffer
	process.Stdout, process.Stderr = &output, &output
	if err := process.Start(); err != nil {
		t.Fatal(err)
	}
	stop := func() {
		if process.Process != nil {
			process.Process.Kill()
			process.Wait()
			process.Process = nil
		}
	}
	t.Cleanup(stop)
	for i := 0; i < 100; i++ {
		probe := exec.Command("redis-cli", "-s", socket, "--user", "sonic-trusted-writer", "PING")
		probe.Env = append(os.Environ(), "REDISCLI_AUTH=rest-local-test-password")
		if reply, err := probe.Output(); err == nil && strings.TrimSpace(string(reply)) == "PONG" {
			return socket, stop
		}
		time.Sleep(10 * time.Millisecond)
	}
	stop()
	t.Fatalf("Redis did not become ready: %s", output.String())
	return "", stop
}

func TestRedisClientWriterPoolPipelineAndReconnect(t *testing.T) {
	socket, _ := fixture(t)
	client := newRedisClient(&redis.Options{Network: "unix", Addr: socket, PoolSize: 3})
	t.Cleanup(func() { client.Close() })
	if err := client.Set("test:key", "before", 0).Err(); err != nil {
		t.Fatal("writer did not authenticate")
	}
	pipeline := client.TxPipeline()
	pipeline.Select(8)
	pipeline.HSet("RESTAPI_TEST", "value", "stored")
	if _, err := pipeline.Exec(); err != nil {
		t.Fatal("writer transaction failed")
	}
	if err := client.ClientKillByFilter("type", "normal", "skipme", "no").Err(); err != nil {
		// The command kills its own connection too. The next call must authenticate again.
		if !strings.Contains(err.Error(), "EOF") {
			t.Fatal("writer connection reset failed")
		}
	}
	// Existing go-redis behavior can return one stale pooled-socket error.
	// A subsequent new connection must still run the named-user login.
	var reconnect error
	for attempt := 0; attempt < 4; attempt++ {
		reconnect = client.Set("test:reconnect", "after", 0).Err()
		if reconnect == nil {
			break
		}
	}
	if reconnect != nil {
		t.Fatal("writer reconnect did not authenticate")
	}
}

func TestRedisClientMissingSecretStops(t *testing.T) {
	socket, _ := fixture(t)
	if err := os.Remove(authRoot + "/local-system/writer.secret"); err != nil {
		t.Fatal(err)
	}
	client := newRedisClient(&redis.Options{Network: "unix", Addr: socket, MaxRetries: -1})
	defer client.Close()
	if err := client.Ping().Err(); err == nil {
		t.Fatal("missing writer state was ignored")
	}
}

func TestRedisClientExistingDialerPreserved(t *testing.T) {
	requireRoot(t)
	called := false
	client := newRedisClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1,
		Dialer: func(ctx context.Context, network, address string) (net.Conn, error) {
			called = true
			if network != "tcp" || address != "127.0.0.1:1" {
				t.Fatal("requested endpoint changed")
			}
			return nil, syscall.ECONNREFUSED
		}})
	defer client.Close()
	client.Ping()
	if !called {
		t.Fatal("original dialer was lost")
	}
}

func TestRedisClientOrdinaryReader(t *testing.T) {
	if os.Getenv("REDIS_REST_READER_TEST") == "1" {
		client := newRedisClient(&redis.Options{Network: "unix", Addr: os.Getenv("REDIS_REST_TEST_SOCKET"), MaxRetries: -1})
		defer client.Close()
		if err := client.Ping().Err(); err != nil {
			t.Fatal("ordinary reader could not connect")
		}
		if value, err := client.Get("test:key").Result(); err != nil || value != "before" {
			t.Fatal("ordinary reader could not read")
		}
		if err := client.Set("test:key", "bad", 0).Err(); err == nil || !strings.Contains(err.Error(), "NOPERM") {
			t.Fatal("ordinary reader write was not denied")
		}
		return
	}
	socket, _ := fixture(t)
	writer := newRedisClient(&redis.Options{Network: "unix", Addr: socket})
	defer writer.Close()
	if err := writer.Set("test:key", "before", 0).Err(); err != nil {
		t.Fatal("writer setup failed")
	}
	binary := filepath.Join(filepath.Dir(socket), "reader.test")
	source, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(binary, source, 0755); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(binary, "-test.run=^TestRedisClientOrdinaryReader$", "-test.v")
	child.Env = append(os.Environ(), "REDIS_REST_READER_TEST=1", "REDIS_REST_TEST_SOCKET="+socket)
	child.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 65534, Gid: 65534}}
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatalf("ordinary-reader child failed: %s", output)
	}
}

func TestRedisClientConfiguredTLSIsPreserved(t *testing.T) {
	requireRoot(t)
	certificateServer := httptest.NewTLSServer(nil)
	configuration := certificateServer.TLS.Clone()
	roots := x509.NewCertPool()
	roots.AddCert(certificateServer.Certificate())
	certificateServer.Close()
	listener, err := tls.Listen("tcp", "127.0.0.1:0", configuration)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	served := make(chan error, 1)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			served <- err
			return
		}
		defer connection.Close()
		connection.SetDeadline(time.Now().Add(2 * time.Second))
		reader := bufio.NewReader(connection)
		for _, expected := range []string{"*1\r\n", "$4\r\n", "ping\r\n"} {
			line, err := reader.ReadString('\n')
			if err != nil {
				served <- err
				return
			}
			if strings.ToLower(line) != expected {
				served <- syscall.EPROTO
				return
			}
		}
		_, err = connection.Write([]byte("+PONG\r\n"))
		served <- err
	}()
	client := newRedisClient(&redis.Options{
		Addr: listener.Addr().String(), MaxRetries: -1,
		DialTimeout: time.Second, ReadTimeout: time.Second,
		TLSConfig: &tls.Config{RootCAs: roots},
	})
	defer client.Close()
	if err := client.Ping().Err(); err != nil {
		t.Fatalf("configured TLS connection failed: %v", err)
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}
