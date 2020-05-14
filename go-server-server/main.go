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
    "time"
    "os"
    "os/signal"
    "syscall"
    "context"
)

func StartHttpServer(handler http.Handler) {
    log.Printf("info: http endpoint started")
    log.Fatal(http.ListenAndServe(":8090", handler))
}


func StartHttpsServer(handler http.Handler) {
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

    server := &http.Server{
        Addr:      ":8081",
        Handler:   handler,
        TLSConfig: tlsConfig,
    }

    log.Printf("info: https endpoint started")

    // Listening should happen in a go-routine to prevent blocking
    go func() {
        log.Fatal(server.ListenAndServeTLS(*sw.ServerCertFlag, *sw.ServerKeyFlag))
    }()

    sigchannel := make(chan os.Signal, 1)
    signal.Notify(sigchannel,
        syscall.SIGTERM,
        syscall.SIGQUIT)

    <-sigchannel
    log.Printf("info: Signal received. Shutting down...")
    if err := server.Shutdown(context.Background()); err != nil {
        log.Printf("HTTP server Shutdown: %v", err)
    } else {
        log.Printf("info: Server shutdown successful!")
    }
    os.Exit(0)
}

func main() {
    iniflags.Parse()

    sw.InitLogging()

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
        go StartHttpsServer(router)
    }

    // infinite loop to keep Goroutines running
    for {
        time.Sleep(1 * time.Second)
    }
}


