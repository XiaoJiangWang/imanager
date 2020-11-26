package encrypt

const (
	CpabeType = "cpabe"
	AesType   = "aes"
)

func Encrypt(text string, encryptType string, role string) (string, error) {
	switch encryptType {
	case CpabeType:
		return encryptWithAttributeBased(text, role)
	case AesType:
		return aesEncrypt(text, key)
	}
	//default
	return encryptWithAttributeBased(text, role)
}

func Decrypt(encryptedData string, encryptType string, role string) (string, error) {
	switch encryptType {
	case CpabeType:
		return decryptWithAttributeBased(encryptedData, role)
	case AesType:
		return aesDecrypt(encryptedData, key)
	}
	return decryptWithAttributeBased(encryptedData, role)
}
