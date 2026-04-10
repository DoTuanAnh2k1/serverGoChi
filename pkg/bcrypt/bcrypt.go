package bcrypt

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

func Encode(password string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		logger.Logger.Errorf("bcrypt: hash password: %v", err)
		return ""
	}
	return string(b)
}

func Matches(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			logger.Logger.Debug("bcrypt: password mismatch")
		} else {
			logger.Logger.Errorf("bcrypt: compare hash: %v", err)
		}
		return false
	}
	return true
}
