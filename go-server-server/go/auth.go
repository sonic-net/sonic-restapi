package restapi

import (
	"log"
	"net/http"
	"strings"
	"strconv"
)

func CommonNameMatch(r *http.Request) bool {
	// During client cert authentication, after the certificate chain is validated by
	// TLS, here we will further check if at least one of the common names in the end-entity certificate
	// matches one of the trusted common names in the server config.

	for _, name := range trustedCertCommonNames {
		is_wildcard := false
		domain := name
		if strings.HasPrefix(name, "*.") {
			if len(name) < 3 {
				log.Printf("warning: Skipping invalid trusted common name %s", name)
				continue;
			}
			is_wildcard = true
			domain = name[1:]  //strip "*" but keep the "." at the beginning
		} else if strings.HasPrefix(name, "*") {
			log.Printf("warning: Skipping invalid trusted common name %s", name)
			continue;
		}
		for _, peercert := range r.TLS.PeerCertificates {
			commonName := peercert.Subject.CommonName
			if is_wildcard {
				// wildcard common name matching
				if len(commonName) > len(domain) && strings.HasSuffix(commonName, domain) {
					log.Printf("info: Wildcard match between common name %s in the client cert and trusted common name %s", commonName, name)
					return true;
				}
			} else {
				if commonName == name {
					log.Printf("info: Exact match with trusted common name %s", name)
					return true;
				}
			}
		}
	}

	commonNames := make([]string, 0)
	for _, peercert := range r.TLS.PeerCertificates {
		commonNames = append(commonNames, strconv.Quote(peercert.Subject.CommonName))
	}
	log.Printf("error: Authentication Failed! None of the common names in the client cert chain" +
			   " matched any of the trusted common names. Client cert common names: %v", commonNames)
	return false;
}