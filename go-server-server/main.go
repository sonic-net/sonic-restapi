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
    "io/ioutil"
    "os"
    "os/signal"
    "syscall"
    "context"
    "sync"
)

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
    wgroup.Wait()
}


