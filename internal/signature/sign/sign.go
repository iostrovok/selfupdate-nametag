package sign

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"os"

	"github.com/pkg/errors"
)

//go:embed key/*
var keyFile embed.FS

type Signature struct {
	key *rsa.PrivateKey
}

func New() (*Signature, error) {
	key, err := loadPrivateKey()
	if err != nil {
		return nil, err
	}

	out := &Signature{
		key: key,
	}

	return out, nil
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	b, err := keyFile.ReadFile("key/id_nametag_key")
	if err != nil {
		return nil, errors.Wrap(err, "signature ReadFile")
	}

	key, err := x509.ParsePKCS1PrivateKey(b)
	return key, errors.Wrap(err, "signature ParsePKCS1PrivateKey")
}

// Sign creates a signature for binaryData using the provided RSA private key.
// Also it creates a signature for the sha256 sum  of binaryData
func (s *Signature) Sign(binaryData []byte) (fileHash, signatureOfHash []byte, err error) {
	if s.key == nil {
		return nil, nil, errors.Errorf("private key is empty. need to call v.loadPrivateKey()")
	}

	msgHash := sha256.New()
	if _, err := msgHash.Write(binaryData); err != nil {
		return nil, nil, errors.Wrap(err, "fileHash msgHash.Write")
	}
	fileHash = msgHash.Sum(nil)

	// Before signing, we need to hash our message. The hash is what we are actually signing
	singHash := sha256.New()
	if _, err := singHash.Write(fileHash); err != nil {
		return nil, nil, errors.Wrap(err, "fileHash msgHash.Write")
	}

	signatureOfHash, err = rsa.SignPSS(rand.Reader, s.key, crypto.SHA256, singHash.Sum(nil), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "fileHash msgHash.Write")
	}

	return fileHash, signatureOfHash, nil
}

// SignFile use Sign to sign the file
func (s *Signature) SignFile(fineName string) (fileHash, signatureOfHash []byte, err error) {
	b, err := os.ReadFile(fineName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "signature SignFile.ReadFile")
	}

	return s.Sign(b)
}
