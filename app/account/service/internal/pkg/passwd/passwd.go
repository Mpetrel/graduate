// Package passwd 对密码进行加密处理
// 使用 bcrypt算法进行密码加密 以及密码验证等
package passwd

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// BCrypt 对密码进行加盐后加密处理
func BCrypt(password string) (encryptPassword string, salt string, err error) {
	salt = uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(password + salt), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	encryptPassword = string(hash)
	return
}

// Verify 检查密码是否正确
func Verify(src string, salt string, target string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(target), []byte(src + salt))
	if err != nil {
		return false
	}
	return true
}