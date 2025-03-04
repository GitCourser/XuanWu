package lib

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 计算字符串的SHA256哈希值
func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
} 