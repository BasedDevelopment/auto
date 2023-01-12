package main

import (
	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/eve/pkg/pki"
	"github.com/BasedDevelopment/eve/pkg/util"
	"github.com/rs/zerolog/log"
)

func checkPKI() {
	if !util.FileExists(keyPath) {
		log.Info().
			Str("path", keyPath).
			Msg("key not found, creating a new one")
		priv := pki.GenKey()
		util.WriteFile(keyPath, priv)
	}
	privBytes := util.ReadFile(keyPath)
	priv := pki.ReadKey(privBytes)

	if !util.FileExists(crtPath) {
		log.Info().
			Str("path", crtPath).
			Msg("crt not found, checking for CSR")
		if !util.FileExists(csrPath) {
			log.Info().
				Str("path", csrPath).
				Msg("CSR not found, creating a new one")
			csr := pki.GenCSR(priv, config.Config.Hostname)
			util.WriteFile(csrPath, csr)
			sum := pki.PemSum(csr)
			log.Info().
				Str("path", csrPath).
				Str("SHA1", sum).
				Msg("CSR written, please send it over to eve to be signed")
			return
		}
		log.Info().
			Str("path", csrPath).
			Msg("CSR found, please send it over to eve to be signed")
		return
	}

	if !util.FileExists(caPath) {
		log.Fatal().
			Str("path", caPath).
			Msg("CA certificate not found, please fetch it from eve")
	}

	crtBytes := util.ReadFile(crtPath)
	crt := pki.ReadCrt(crtBytes)
	caBytes := util.ReadFile(caPath)
	ca := pki.ReadCrt(caBytes)

	if err := pki.VerifyCrt(ca, crt); err != nil {
		log.Fatal().
			Err(err).
			Str("cert path", crtPath).
			Str("ca path", caPath).
			Msg("certificate verification failed")
	}

	crtSum := pki.PemSum(crtBytes)
	caSum := pki.PemSum(caBytes)

	log.Info().
		Str("cert path", crtPath).
		Str("ca path", caPath).
		Str("crt SHA1", crtSum).
		Str("ca SHA1", caSum).
		Msg("certificate verification succeeded")
}
