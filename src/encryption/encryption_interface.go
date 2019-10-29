package encryption

type IEncryptionMethod interface {
	EncryptionMethod() string

	SetEncryptionArgs(map[string]string)

	EncryptData([]byte) []byte

	DecryptData([]byte) []byte
}
