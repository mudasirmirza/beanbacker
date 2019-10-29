package encryption

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type KMSEncryption struct {
	name        string
	EncryptionArgs map[string]string
}

func (e KMSEncryption) EncryptionMethod() string {
	return e.name
}

func (e *KMSEncryption) SetEncryptionArgs(encryptionArgs map[string]string) {
	e.EncryptionArgs = encryptionArgs
}

func (e KMSEncryption) getKMSClient() *kms.KMS {
	sess, _ := session.NewSession()
	return kms.New(sess)
}

func (e KMSEncryption) EncryptData(data []byte) []byte {
	result, _ := e.getKMSClient().Encrypt(&kms.EncryptInput{
		KeyId: aws.String(e.EncryptionArgs["kmsKeyArn"]),
		Plaintext: []byte(e.EncryptionArgs["passphrase"]),
	})
	return AESCipherEncryption{EncryptionArgs: map[string]string{"passphrase": string(result.CiphertextBlob)}}.EncryptData(data)
}

func (e KMSEncryption) DecryptData(data []byte) []byte {
	result, _ := e.getKMSClient().Decrypt(&kms.DecryptInput{
		CiphertextBlob: data,
	})
	return AESCipherEncryption{EncryptionArgs: map[string]string{"passphrase": string(result.Plaintext)}}.DecryptData(data)
}
