package main

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"
)

func TestGenSelfCert(t *testing.T) {
	err := GenSelfSignedCert("server.pem", "server.crt")
	if err != nil {
		t.Error(err)
	}
	//check that our private key file exists
	if _, err := os.Stat("server.pem"); os.IsNotExist(err) {
		t.Error("Could not locate self signed certificate private key after generation")
	}
	//check that we can decode it
	privKeyPEM, err := ioutil.ReadFile("server.pem")
	if err != nil {
		t.Error(err)
	}
	privKeyPEMDecoded, _ := pem.Decode(privKeyPEM)
	if privKeyPEMDecoded.Type != "EC PRIVATE KEY" {
		t.Errorf("Expected private key type: %s Actual: %s", "EC PRIVATE KEY", privKeyPEMDecoded.Type)
	}
	_, err = x509.ParseECPrivateKey(privKeyPEMDecoded.Bytes)
	if err != nil {
		t.Error(err)
	}

	//check that our certificate exists
	if _, err := os.Stat("server.crt"); os.IsNotExist(err) {
		t.Error("Could not locate self signed certificate after generation")
	}
	//check that we can decode it
	certPEM, err := ioutil.ReadFile("server.crt")
	if err != nil {
		t.Error(err)
	}
	certPEMDecoded, _ := pem.Decode(certPEM)
	if certPEMDecoded.Type != "CERTIFICATE" {
		t.Errorf("Expected certificate type: %s Actual: %s", "CERTIFICATE", certPEMDecoded.Type)
	}
	cert, err := x509.ParseCertificate(certPEMDecoded.Bytes)
	if err != nil {
		t.Error(err)
	}
	if cert.Issuer.Organization[0] != "Jumpcloud" {
		t.Errorf("Expected issuing organization: %s Actual: %s", "Jumpcloud", cert.Issuer.Organization[0])
	}

	//delete cert and key
	os.Remove("server.pem")
	os.Remove("server.crt")
}

func TestCheckCertExists(t *testing.T) {
	err := GenSelfSignedCert("server.pem", "server.crt")
	if err != nil {
		t.Error(err)
	}
	exist := CheckCertExists("server.pem", "server.crt")
	if !exist {
		t.Error("Could not find certificate and key after generation")
	}
	//delete key
	os.Remove("server.pem")
	exist = CheckCertExists("server.pem", "server.crt")
	if exist {
		t.Error("Found both although key has been deleted")
	}

	//delete cert
	os.Remove("server.crt")
	exist = CheckCertExists("server.pem", "server.crt")
	if exist {
		t.Error("Found both although both key and certificate been deleted")
	}
}
