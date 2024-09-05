package verify

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"encoding/base64"

	"github.com/pkg/errors"
)

//go:embed key/*
var keyFile embed.FS

type Verifier struct {
	key *rsa.PublicKey
}

func New() (*Verifier, error) {
	key, err := loadPublicKey()
	if err != nil {
		return nil, err
	}

	out := &Verifier{
		key: key,
	}

	return out, nil
}

func loadPublicKey() (*rsa.PublicKey, error) {
	publicBytes, err := keyFile.ReadFile("key/id_nametag_key_pub")
	if err != nil {
		return nil, errors.Wrap(err, "signature ReadFile")
	}

	key, err := x509.ParsePKCS1PublicKey(publicBytes)
	return key, errors.Wrap(err, "signature ParsePublicKey")
}

// Verify checks the signature of binaryData against a signature.
func (v *Verifier) Verify(binaryData, signature string) error {
	if v.key == nil {
		return errors.Errorf("public key is empty. need to call v.loadPublicKey()")
	}

	binaryDataB, err := base64.URLEncoding.DecodeString(binaryData)
	if err != nil {
		return errors.Wrap(err, "binaryData DecodeString")
	}

	signatureB, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		return errors.Wrap(err, "signature DecodeString")
	}

	msgHash := sha256.New()
	if _, err := msgHash.Write(binaryDataB); err != nil {
		return errors.Wrap(err, "msgHash.Write")
	}

	return rsa.VerifyPSS(v.key, crypto.SHA256, msgHash.Sum(nil), signatureB, nil)
}
