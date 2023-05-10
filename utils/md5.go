package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	tempStr := h.Sum(nil)
	return hex.EncodeToString(tempStr)
}
//大写
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

//加密
func MakePassword(plainpwd, salt string) string {
	return Md5Encode(plainpwd + salt)
}
//解密,判断数据库的密码解密之后与用户输入的password是否一直
func ValiPassword(plainpwd, salt string, password string) bool {
	return Md5Encode(plainpwd+salt) == password
}
