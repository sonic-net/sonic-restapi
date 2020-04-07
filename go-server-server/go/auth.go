package mseeserver

import (
    "log"
    "net/http"
)

func CommonNameMatch(r *http.Request) bool {
    //FIXME : in the authentication of client certificate,  after the certificate chain is validated by 
    // TLS, here we will futher check if the common name of the end-entity certificate is in the trusted 
	// common name list of the server config. A more strict check may be added here later.
	for _, peercert := range r.TLS.PeerCertificates {
		commonName := peercert.Subject.CommonName
		for _, name := range trustedertCommonNames {
			if commonName == name {
				return true;
			}
		}
		log.Printf("info: CommonName in the client cert: %s", commonName)
	}

    log.Printf("error: Authentication Fail! None of the CommonNames in the client cert are found in trusted common names")
    return false;
}