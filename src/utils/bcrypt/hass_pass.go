package bcrypt

import (
	"errors"
	"serverGoChi/src/logger"

	"golang.org/x/crypto/bcrypt"
)

func Encode(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		logger.Logger.Error("Cant hash password")
		return ""
	}
	return string(bytes)
}

func Matches(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			logger.Logger.Info("Password does not match the hash")
		} else {
			logger.Logger.Error("Error comparing hash and password: ", err)
		}
		return false
	}
	return true
}
