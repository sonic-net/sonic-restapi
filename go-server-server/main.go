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
    log.Fatal(server.ListenAndServeTLS(*sw.ServerCertFlag, *sw.ServerKeyFlag))
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


