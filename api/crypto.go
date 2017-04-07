package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

type SSLConfig struct {
	CertFile  string //path to X509 certificate chain
	KeyFile   string //path to key file
	Port      uint   //port to listen on for HTTPS connections
	Exclusive bool   //if true, do not allow non-HTTPS connections
}

//Generates a self-signed X509 certificate and keypair using ECDSA with NIST P-256 curve
//
//Writes the x509 certificate and private key to certDest and keyDest respectively as PEM encoded DER
func GenSelfSignedCert(keyDest string, certDest string) error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1337),
		Subject: pkix.Name{
			Country:            []string{"United States"},
			Organization:       []string{"Jumpcloud"},
			OrganizationalUnit: []string{"Interviews"},
		},
		Issuer: pkix.Name{
			Country:            []string{"United States"},
			Organization:       []string{"Jumpcloud"},
			OrganizationalUnit: []string{"Jumpcloud Root CA"},
			Locality:           []string{"Boulder"},
			Province:           []string{"Colorado"},
			SerialNumber:       "255",
			CommonName:         "255",
		},
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
		PublicKeyAlgorithm:    x509.ECDSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 7),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	pubKey := privKey.Public()
	caCertDER, err := x509.CreateCertificate(rand.Reader, cert, cert, pubKey, privKey) //this is now DER encoded
	if err != nil {
		return err
	}
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER}) //need to wrap it as PEM
	err = ioutil.WriteFile(certDest, caCertPEM, 0644)
	if err != nil {
		return err
	}

	privKeyDER, err := x509.MarshalECPrivateKey(privKey)                                    //now DER encoded
	pemPrivKey := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyDER}) //we need to wrap it as PEM
	err = ioutil.WriteFile(keyDest, pemPrivKey, 0644)
	if err != nil {
		return err
	}
	return nil
}

//Check that both the certificate and keyfile exist at the given path
func CheckCertExists(keyFile string, certFile string) bool {
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return false
	} else if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return false
	}
	return true
}
