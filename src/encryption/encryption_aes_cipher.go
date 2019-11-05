package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

type AESCipherEncryption struct {
	name           string
	EncryptionArgs map[string]string
}

func (e AESCipherEncryption) EncryptionMethod() string {
	return e.name
}

func (e *AESCipherEncryption) SetEncryptionArgs(encryptionArgs map[string]string) {
	e.EncryptionArgs = encryptionArgs
}

// func to create a random hash
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Func to encrypt text
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func (e AESCipherEncryption) EncryptData(data []byte) EncryptionOutput {
	passphrase := e.EncryptionArgs["passphrase"]
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return EncryptionOutput{EncryptedData: ciphertext}
}

// Func to decrypt text
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func (e AESCipherEncryption) DecryptData(data []byte) []byte {
	passphrase := e.EncryptionArgs["passphrase"]
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
