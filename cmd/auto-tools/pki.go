package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/rs/zerolog/log"
)

func createKey() {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate key")
	}
	pks8Priv, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal private key")
	}
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pks8Priv,
	})
	err = ioutil.WriteFile(keyFile, privPem, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write private key")
	}
	log.Info().Msg("Private key generated")
}

func createCSR() {
	privKeyFile, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read private key")
	}
	block, _ := pem.Decode(privKeyFile)
	privPem := block.Bytes
	priv, err := x509.ParsePKCS8PrivateKey(privPem)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse private key")
	}

	csrTemp := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: config.Config.Hostname,
		},
		DNSNames: []string{config.Config.Hostname},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, csrTemp, priv)
	csrPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create certificate request")
	}
	err = ioutil.WriteFile(csrFile, csrPem, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write certificate request")
	}
}

func validateCert() {
	// Open and parse cert
	certPem, err := ioutil.ReadFile(crtFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read certificate")
	}
	cert, err := x509.ParseCertificate(certPem)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse certificate")
	}

	// Open and parse CA
	caPem, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read CA")
	}
	ca, err := x509.ParseCertificate(caPem)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse CA")
	}

	// Verify cert
	if cert.Subject.CommonName != config.Config.Hostname {
		log.Fatal().Msg("Certificate common name does not match hostname")
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	if _, err := cert.Verify(x509.VerifyOptions{Roots: pool}); err != nil {
		log.Fatal().Err(err).Msg("Failed to verify certificate")
	}
	if err := cert.CheckSignatureFrom(ca); err != nil {
		log.Fatal().Err(err).Msg("Failed to check signature of certificate")
	}

	log.Info().Msg("Certificate validated")
}

func certSum() {
	certBytes, err := ioutil.ReadFile(crtFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read certificate")
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse certificate")
	}
	certShaBytes := sha256.Sum256(cert.Raw)
	certSha := hex.EncodeToString(certShaBytes[:])
	log.Info().
		Str("fingerprint", certSha).
		Msg("Certificate fingerprint")
}

func csrSum() {
	csrPemBytes, err := ioutil.ReadFile(csrFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read csr")
	}
	csrPem, _ := pem.Decode(csrPemBytes)
	csr, err := x509.ParseCertificateRequest(csrPem.Bytes)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse csr")
	}
	csrShaBytes := sha256.Sum256(csr.Raw)
	csrSha := hex.EncodeToString(csrShaBytes[:])
	log.Info().
		Str("fingerprint", csrSha).
		Msg("CSR fingerprint")
}

func caSum() {
	caPemBytes, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read CA")
	}
	pem, _ := pem.Decode(caPemBytes)
	ca, err := x509.ParseCertificate(pem.Bytes)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse CA")
	}
	caShaBytes := sha256.Sum256(ca.Raw)
	caSha := hex.EncodeToString(caShaBytes[:])
	log.Info().
		Str("fingerprint", caSha).
		Msg("CA fingerprint")
}
