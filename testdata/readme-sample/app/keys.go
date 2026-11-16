package app

import (
	"crypto/mlkem"
	"crypto/rand"
	"crypto/rsa"
)

func signingKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 1024)
}

func sessionKey() (*mlkem.DecapsulationKey768, error) {
	return mlkem.GenerateKey768()
}
