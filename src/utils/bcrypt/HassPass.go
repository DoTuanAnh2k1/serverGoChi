package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
	"serverGoChi/src/log"
)

func Encode(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Logger.Error("Cant hash password")
		return ""
	}
	return string(bytes)
}

func Matches(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Logger.Error("Cant check password")
		return false
	}
	return true
}
