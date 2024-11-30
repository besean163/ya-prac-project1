package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func main() {
	if err := createKeys(); err != nil {
		log.Fatal(err)
	}
}

func createKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var publicKeyPEM bytes.Buffer
	pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	file, err := os.OpenFile("public_key.pem", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	_, err = file.WriteString(publicKeyPEM.String())
	if err != nil {
		return err
	}

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	file, err = os.OpenFile("private_key.pem", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	_, err = file.WriteString(privateKeyPEM.String())
	if err != nil {
		return err
	}
	return nil
}
