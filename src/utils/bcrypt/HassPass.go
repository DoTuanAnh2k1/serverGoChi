package bcrypt

import (
	"errors"
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
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Logger.Info("Password does not match the hash")
		} else {
			log.Logger.Error("Error comparing hash and password: ", err)
		}
		return false
	}
	return true
}
