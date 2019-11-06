package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	keystore "github.com/dgrijalva/jwt-go"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	privateKeyPath = filepath.FromSlash("keys/app.rsa")
	publicKeyPath  = filepath.FromSlash("keys/app.rsa.pub")
	PrivateRSAKey  *rsa.PrivateKey
	publicRSAKey   *rsa.PublicKey
)

// Generate/read rsa keys
func InitJWT() {
	if !FileExists(privateKeyPath) {
		// If there is no private key file, generate new key pair and save
		generateRSAKey()
	}

	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	CheckError(err)

	PrivateRSAKey, err = keystore.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	CheckError(err)

	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	CheckError(err)

	publicRSAKey, err = keystore.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	CheckError(err)
}

func generateRSAKey() {
	// Create RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	CheckError(err)

	// Create directory to store keys
	if !DirectoryExists("keys") {
		err = os.Mkdir("keys", 0777)
		CheckError(err)
	}

	// Store public and private keys in separate files
	savePrivatePEMKey(privateKeyPath, privateKey)
	savePublicPEMKey(publicKeyPath, privateKey.PublicKey)
}

func savePrivatePEMKey(filepath string, privateKey *rsa.PrivateKey) {
	saveFile, err := os.Create(filepath)
	CheckError(err)
	defer saveFile.Close()

	var pemKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(saveFile, pemKey)
	CheckError(err)

	// Set file permissions to read/write for owner only
	err = os.Chmod(filepath, 0777)
	CheckError(err)
}

func savePublicPEMKey(filepath string, publicKey rsa.PublicKey) {
	bytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	CheckError(err)

	saveFile, err := os.Create(filepath)
	CheckError(err)
	defer saveFile.Close()

	var pemKey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}

	err = pem.Encode(saveFile, pemKey)
	CheckError(err)

	// Set file permissions to read/write for owner only
	err = os.Chmod(filepath, 0777)
	CheckError(err)
}

// CreateJWT will create a JWT signed by a private RSA key
func CreateJWT(user User) string {
	// Create Square.jose signing key
	signingKey := jose.SigningKey{Algorithm: jose.RS256, Key: PrivateRSAKey}

	// Create Square.jose RSA signer
	signerOptions := jose.SignerOptions{}
	signerOptions.WithType("JWT")
	rsaSigner, err := jose.NewSigner(signingKey, &signerOptions)

	// Create an instance of JWT Builder that uses the RSA signer
	builder := jwt.Signed(rsaSigner)

	// Create an instance of CustomClaim
	customClaim := CustomClaim{
		Claims: &jwt.Claims{
			Issuer:   "test",
			Subject:  user.Username,
			ID:       user.ID,
			Audience: jwt.Audience{"test"},
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Expiry:   jwt.NewNumericDate(time.Now().AddDate(0, 0, 1)), // Set to expire after 1 day
		},
	}

	// Add custom claim to the builder
	builder = builder.Claims(customClaim)

	// Sign with RSA key and return compact JWT
	rawJWT, err := builder.CompactSerialize()
	CheckError(err)

	return rawJWT
}
