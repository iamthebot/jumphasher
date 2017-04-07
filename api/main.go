package main

import (
	"flag"
	"github.com/iamthebot/jumphasher/common"
	"log"
	"runtime"
)

func main() {
	var sslmode string
	var port uint
	var delay uint
	var sslcfg SSLConfig
	var concurrency uint

	flag.StringVar(&sslmode, "sslmode", "hybrid", "'hybrid' (serve both HTTP and HTTPS), 'exclusive' (HTTPS only), or 'disabled' (HTTP only)")
	flag.UintVar(&port, "port", 80, "port to use for HTTP")
	flag.UintVar(&sslcfg.Port, "sslport", 443, "port to use for HTTPS")
	flag.UintVar(&delay, "delay", 5, "number of seconds to delay hashing")
	flag.StringVar(&sslcfg.CertFile, "sslcert", "server.crt", "path to server X509 SSL certificate in PEM format. If a certificate/key pair is not found and SSL is enabled a self-signed one will be generated in this file")
	flag.StringVar(&sslcfg.KeyFile, "sslkey", "server.pem", "path to server private key. If a certificate/key pair is not found and SSL is enabled, an elliptic key based on NIST P-256 will be generated in this file")
	flag.UintVar(&concurrency, "concurrency", uint(runtime.NumCPU()), "target concurrency for API server and data structures")
	flag.Parse()
	if port > 65535 {
		log.Fatalf("Port %d exceeds max port number 65535", port)
	} else if sslcfg.Port > 65535 {
		log.Fatalf("HTTPS Port %d exceeds max port number 65535", sslcfg.Port)
	}
	switch sslmode {
	case "hybrid":
		sslcfg.Exclusive = false
	case "exclusive":
		sslcfg.Exclusive = true
	case "disabled":
		sslcfg.Exclusive = false
	default:
		log.Fatalf("Unknown sslmode '%s'", sslmode)
	}
	var engine *APIEngine
	var err error
	if sslmode != "disabled" {
		exists := CheckCertExists(sslcfg.KeyFile, sslcfg.CertFile)
		if !exists {
			GenSelfSignedCert(sslcfg.KeyFile, sslcfg.CertFile)
		}
		engine, err = NewAPIEngine(int(concurrency), jumphasher.HashTypeSHA512, &sslcfg, int(port), int(delay))
	} else {
		engine, err = NewAPIEngine(int(concurrency), jumphasher.HashTypeSHA512, nil, int(port), int(delay))
	}
	if err != nil {
		log.Fatal(err)
	}
	engine.Start()
}
