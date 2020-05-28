package main

import (
    // WARNING!
    // Change this to a fully-qualified import path
    // once you place this file into your project.
    // For example,
    //
    //    sw "github.com/myname/myrepo/go"
    //
    "github.com/vharitonsky/iniflags"
    sw "go-server-server/go"
    "log"
    "net/http"
    "crypto/tls"
    "crypto/x509"
    "crypto/md5"
    "io"
    "io/ioutil"
    "os"
    "os/signal"
    "syscall"
    "context"
    "sync"
    "time"
    "bytes"
)

const CERT_MONITOR_FREQUENCY = 3600 * time.Second

func StartHttpServer(handler http.Handler) {
    log.Printf("info: http endpoint started")
    log.Fatal(http.ListenAndServe(":8090", handler))
}

func StartHttpsServer(handler http.Handler, messenger <-chan int, wgroup *sync.WaitGroup) {
    defer wgroup.Done()
    clientCert, err := ioutil.ReadFile(*sw.ClientCertFlag)
    if err != nil {
        log.Fatalf("error: couldn't open client cert file, %s", err)
    }
    clientCertPool := x509.NewCertPool()
    clientCertPool.AppendCertsFromPEM(clientCert)

    // Setup HTTPS client cert the server trust and validation policy
    tlsConfig := &tls.Config{
        ClientCAs: clientCertPool,
        // NoClientCert
        // RequestClientCert
        // RequireAnyClientCert
        // VerifyClientCertIfGiven
        // RequireAndVerifyClientCert
        ClientAuth: tls.RequireAndVerifyClientCert,
        MinVersion: tls.VersionTLS12,
    }

    tlsConfig.BuildNameToCertificate()
    for {
        server := &http.Server{
            Addr:      ":8081",
            Handler:   handler,
            TLSConfig: tlsConfig,
        }

        log.Printf("info: https endpoint started")

        // Listening should happen in a go-routine to prevent blocking
        go func() {
            if err := server.ListenAndServeTLS(*sw.ServerCertFlag, *sw.ServerKeyFlag); err != nil {
                log.Println(err)
            }
        }()

        value := <-messenger
        log.Printf("info: HTTPS Signal received. Shutting down...")
        if err := server.Shutdown(context.Background()); err != nil {
            log.Printf("trace: HTTPS server Shutdown: %v", err)
        } else {
            log.Printf("info: HTTPS Server shutdown successful!")
        }
        switch value {
        case 0:
            log.Printf("info: Terminating...")
            return
        }
    }
}

func signal_handler(messenger chan<- int, wgroup *sync.WaitGroup) {
    defer wgroup.Done()
    sigchannel := make(chan os.Signal, 1)
    signal.Notify(sigchannel,
        syscall.SIGTERM,
        syscall.SIGKILL,
        syscall.SIGQUIT)

    <-sigchannel
    messenger <- 0
    return
}

func monitor_certs(messenger chan<- int, wgroup *sync.WaitGroup) {
    defer wgroup.Done()
    client_cert, _ := os.Open(*sw.ClientCertFlag)
    prev_client_cert_hash := md5.New()
    io.Copy(prev_client_cert_hash, client_cert)
    log.Printf("trace: MD5 checksum of %s is %x", client_cert.Name(), prev_client_cert_hash.Sum(nil))
    client_cert.Close()

    server_cert, _ := os.Open(*sw.ServerCertFlag)
    prev_server_cert_hash := md5.New()
    io.Copy(prev_server_cert_hash, server_cert)
    log.Printf("trace: MD5 checksum of %s is %x", server_cert.Name(), prev_server_cert_hash.Sum(nil))
    server_cert.Close()

    server_key, _ := os.Open(*sw.ServerKeyFlag)
    prev_server_key_hash := md5.New()
    io.Copy(prev_server_key_hash, server_key)
    log.Printf("trace: MD5 checksum of %s is %x", server_key.Name(), prev_server_key_hash.Sum(nil))
    server_key.Close()

    time.Sleep(CERT_MONITOR_FREQUENCY)

    for {
        reload := false
        client_cert, _ := os.Open(*sw.ClientCertFlag)
        client_cert_hash := md5.New()
        io.Copy(client_cert_hash, client_cert)    
        log.Printf("trace: MD5 checksum of %s is %x", client_cert.Name(), client_cert_hash.Sum(nil))
        if bytes.Compare(client_cert_hash.Sum(nil), prev_client_cert_hash.Sum(nil)) != 0 {
            log.Printf("info: MD5 checksum of %s changed from %x to %x", client_cert.Name(), prev_client_cert_hash.Sum(nil), client_cert_hash.Sum(nil))
            reload = true
        }
        prev_client_cert_hash = client_cert_hash
        client_cert.Close()

        server_cert, _ := os.Open(*sw.ServerCertFlag)
        server_cert_hash := md5.New()
        io.Copy(server_cert_hash, server_cert)    
        log.Printf("trace: MD5 checksum of %s is %x", server_cert.Name(), server_cert_hash.Sum(nil))
        if bytes.Compare(server_cert_hash.Sum(nil), prev_server_cert_hash.Sum(nil)) != 0 {
            log.Printf("info: MD5 checksum of %s changed from %x to %x", server_cert.Name(), prev_server_cert_hash.Sum(nil), server_cert_hash.Sum(nil))
            reload = true
        }
        prev_server_cert_hash = server_cert_hash
        server_cert.Close()

        server_key, _ := os.Open(*sw.ServerKeyFlag)
        server_key_hash := md5.New()
        io.Copy(server_key_hash, server_key)    
        log.Printf("trace: MD5 checksum of %s is %x", server_key.Name(), server_key_hash.Sum(nil))
        if bytes.Compare(server_key_hash.Sum(nil), prev_server_key_hash.Sum(nil)) != 0 {
            log.Printf("info: MD5 checksum of %s changed from %x to %x", server_key.Name(), prev_server_key_hash.Sum(nil), server_key_hash.Sum(nil))
            reload = true
        }
        prev_server_key_hash = server_key_hash
        server_key.Close()

        if reload == true {
            log.Printf("info: Certs have rolled! Reload needed!")
            messenger <- 1
        }
        time.Sleep(CERT_MONITOR_FREQUENCY)
    }
}

func main() {
    iniflags.Parse()

    sw.InitLogging()
    var wgroup sync.WaitGroup
    var messenger = make(chan int, 1)

    log.Printf("info: server started")

    sw.Initialise()
    router := sw.NewRouter()

    if (!*sw.HttpFlag && !*sw.HttpsFlag) {
        log.Fatal("Both http and http endpoints are disabled.")
    }

    if (*sw.HttpFlag) {
        go StartHttpServer(router)
    }

    if (*sw.HttpsFlag) {
        wgroup.Add(1)
        go StartHttpsServer(router, messenger, &wgroup)
    }

    wgroup.Add(1)
    go signal_handler(messenger, &wgroup)

    wgroup.Add(1)
    go monitor_certs(messenger, &wgroup)

    wgroup.Wait()
}
