package encryption

type IEncryptionMethod interface {
	EncryptionMethod() string

	SetEncryptionArgs(map[string]string)

	EncryptData([]byte) EncryptionOutput

	DecryptData([]byte) []byte
}
