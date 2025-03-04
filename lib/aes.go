package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"os"
)

//加密过程：
//  1、处理数据，对数据进行填充，采用PKCS7（当密钥长度不够时，缺几位补几个几）的方式。
//  2、对数据进行加密，采用AES加密方法中CBC加密模式
//  3、对得到的加密数据，进行base64加密，得到字符串
// 解密过程相反

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	if unPadding > length {
		return nil, errors.New("填充长度错误")
	}
	if unPadding == 0 {
		return nil, errors.New("填充长度不能为0")
	}
	// 验证填充是否合法
	for i := length - unPadding; i < length; i++ {
		if data[i] != byte(unPadding) {
			return nil, errors.New("非法的PKCS7填充")
		}
	}
	return data[:(length - unPadding)], nil
}

// AesEncrypt 加密
func AesEncrypt(data []byte, key []byte) ([]byte, error) {
	//创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//判断加密快的大小
	blockSize := block.BlockSize()
	//填充
	encryptBytes := pkcs7Padding(data, blockSize)
	//初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	//使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	//执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// AesDecrypt 解密
func AesDecrypt(data []byte, key []byte) ([]byte, error) {
	//创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//初始化解密数据接收切片
	crypted := make([]byte, len(data))
	//执行解密
	blockMode.CryptBlocks(crypted, data)
	//去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}

// EncryptByAes Aes加密 后 base64 再加
func EncryptByAes(data []byte) (string, error) {
	// 16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
	// key不能泄露
	PwdKey := generateKey()
	res, err := AesEncrypt(data, PwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

// DecryptByAes Aes 解密
func DecryptByAes(data string) ([]byte, error) {
	if data == "" {
		return nil, errors.New("加密数据不能为空")
	}
	
	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.New("base64解码失败：" + err.Error())
	}
	
	if len(dataByte) == 0 || len(dataByte)%aes.BlockSize != 0 {
		return nil, errors.New("无效的加密数据长度")
	}
	
	// key不能泄露
	PwdKey := generateKey()
	return AesDecrypt(dataByte, PwdKey)
}

// 生成16位字符串
func generateKey() []byte {
	str, _ := os.Executable()
	key := make([]byte, 0, 16)
	if len(str) > 16 {
		key = append(key, str[:16]...)
	} else {
		key = append(key, str...)
		remain := 16 - len(str)
		for i := 0; i < remain; i++ {
			key = append(key, 'A')
		}
	}

	// println(string(key))
	return []byte(key)
}
