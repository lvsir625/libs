package crypto

import (
	"crypto/aes"
	"bytes"
	"crypto/cipher"
	"errors"
)

//AES ECB模式的加密解密
type AesECBCipher struct {
	//128 192  256位的其中一个 长度 对应分别是 16 24  32字节长度
	key       []byte
	block     cipher.Block
	blockSize int
}

func (aec *AesECBCipher) PCSK5Padding(src []byte) []byte {
	//填充个数
	paddingCount := aes.BlockSize - len(src)%aes.BlockSize
	if paddingCount == 0 {
		return src
	} else {
		//填充数据
		return append(src, bytes.Repeat([]byte{byte(paddingCount)}, paddingCount)...)
	}
}

//unpadding
func (aec *AesECBCipher) PCSK5UnPadding(src []byte) ([]byte, error) {
	paddingCount := int(src[len(src)-1])
	if paddingCount > len(src) {
		return nil, errors.New("PCSK5UnPadding err")
	}
	src = src[:len(src)-paddingCount]
	return src, nil
}

func (aec *AesECBCipher) Encrypt(src []byte) ([]byte) {
	//padding
	src = aec.PCSK5Padding(src)
	//返回加密结果
	encryptData := make([]byte, len(src))
	//分组分块加密
	for index := 0; index < len(src); index += aec.blockSize {
		aec.block.Encrypt(encryptData[index:], src[index:index+aec.blockSize])
	}
	return encryptData
}
func (aec *AesECBCipher) Decrypt(src []byte) ([]byte, error) {
	if len(src) == 0 || len(src) % aec.blockSize != 0 {
		return nil, errors.New("aes source data length wrong")
	}
	//返回加密结果
	decryptData := make([]byte, len(src))

	//分组分块加密
	for index := 0; index < len(src); index += aec.blockSize {
		aec.block.Decrypt(decryptData[index:], src[index:index+aec.blockSize])
	}

	return aec.PCSK5UnPadding(decryptData)
}

func CreateAesECBCipher(key []byte, blockSize int) (*AesECBCipher, error) {
	//key只能是 16 24 32长度
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &AesECBCipher{
		key: key,
		block: block,
		blockSize: blockSize,
	}, nil
}