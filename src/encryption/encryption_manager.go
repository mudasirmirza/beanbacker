package encryption

var AvailableEncryptionMethods map[string]IEncryptionMethod

func GetAvailableEncryptionMethodNames() []string {
	keys := make([]string, 0, len(AvailableEncryptionMethods))
	for key := range AvailableEncryptionMethods {
		keys = append(keys, key)
	}
	return keys
}

func register(encryptionMethod IEncryptionMethod) {
	if AvailableEncryptionMethods == nil {
		AvailableEncryptionMethods = make(map[string]IEncryptionMethod)
	}
	AvailableEncryptionMethods[encryptionMethod.EncryptionMethod()] = encryptionMethod
}

func Init() {
	register(&AESCipherEncryption{name: "AESCipherEncryption"})
	register(&KMSEncryption{name: "KMSEncryption"})
}
