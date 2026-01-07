package restapi

import (
	"log"
	"net/http"
	"strings"
)

func CommonNameMatch(r *http.Request) bool {
    //FIXME : in the authentication of client certificate,  after the certificate chain is validated by 
    // TLS, here we will futher check if at least one of the common names in the end-entity certificate
	// matches one of the trusted common names of the server config. A more strict check may be added here later.
	for _, peercert := range r.TLS.PeerCertificates {
		commonName := peercert.Subject.CommonName
		for _, name := range trustedCertCommonNames {
			if strings.HasPrefix(name, "*") {
				// wildcard common name match
				domain := name[1:]  //strip "*"
				if strings.HasSuffix(commonName, domain) {
					log.Printf("info: CommonName %s in the client cert matches trusted wildcard common name %s", commonName, name)
					return true;
				}
			} else if commonName == name {
				return true;
			}
		}
		log.Printf("info: CommonName in the client cert: %s", commonName)
	}

    log.Printf("error: Authentication Fail! None of the CommonNames in the client cert match any of the trusted common names")
    return false;
}