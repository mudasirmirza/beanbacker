package encryption

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type KMSEncryption struct {
	name           string
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

func (e KMSEncryption) generateDataKey() *kms.GenerateDataKeyOutput {
	kmsClient := e.getKMSClient()
	keySpec := "AES_256"
	dataKey, _ := kmsClient.GenerateDataKey(&kms.GenerateDataKeyInput{KeyId: aws.String(e.EncryptionArgs["kmsKeyArn"]), KeySpec: &keySpec})
	return dataKey
}

func (e KMSEncryption) EncryptData(data []byte) EncryptionOutput {
	dataKey := e.generateDataKey()
	return EncryptionOutput{EncryptedData: AESCipherEncryption{EncryptionArgs: map[string]string{"passphrase": string(dataKey.Plaintext)}}.EncryptData(data).EncryptedData, EncryptedDataKey: dataKey.CiphertextBlob}
}

func (e KMSEncryption) DecryptData(data []byte) []byte {
	kmsClient := e.getKMSClient()
	dataKey, _ := kmsClient.Decrypt(&kms.DecryptInput{CiphertextBlob: []byte(e.EncryptionArgs["encryptedDataKey"])})
	return AESCipherEncryption{EncryptionArgs: map[string]string{"passphrase": string(dataKey.Plaintext)}}.DecryptData(data)
}
