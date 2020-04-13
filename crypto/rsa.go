package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type RsaCipher struct {
	encryptBlock *pem.Block
	decryptBlock *pem.Block
	pub_key *rsa.PublicKey
	pri_key *rsa.PrivateKey
}

// 加密
func (rc *RsaCipher) Encrypt(origData []byte) ([]byte, error) {
	if rc.pub_key == nil {
		pubInterface, err := x509.ParsePKIXPublicKey(rc.encryptBlock.Bytes) //解析pem.Decode（）返回的Block指针实例
		if err != nil {
			return nil, err
		}

		pub, ok := pubInterface.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("x509.ParsePKIXPublicKey result not a PublicKey interface")
		}

		rc.pub_key = pub
	}

	return rsa.EncryptPKCS1v15(rand.Reader, rc.pub_key, origData) //RSA算法加密
}

// 解密
func (rc *RsaCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if rc.pri_key == nil {
		priv, err := x509.ParsePKCS1PrivateKey(rc.decryptBlock.Bytes) //解析pem.Decode（）返回的Block指针实例
		if err != nil {
			return nil, err
		}

		rc.pri_key = priv
	}

	return rsa.DecryptPKCS1v15(rand.Reader, rc.pri_key, ciphertext) //RSA算法解密
}

func CreateRsaCipher(privateKey, publicKey []byte) (*RsaCipher, error) {
	encryptBlock, _ := pem.Decode(publicKey) //将密钥解析成公钥实例
	if encryptBlock == nil {
		return nil, errors.New("public key error")
	}

	decryptBlock, _ := pem.Decode(privateKey) //将密钥解析成私钥实例
	if decryptBlock == nil {
		return nil, errors.New("private key error!")
	}

	return &RsaCipher{
		decryptBlock:decryptBlock,
		encryptBlock:encryptBlock,
	}, nil
}