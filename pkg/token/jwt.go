package token

import (
	"errors"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

const prefixKey = "Basic "

func getSecretKey() []byte {
	key := config.GetJwtConfig().SecretKey
	if key == "" {
		logger.Logger.Warn("token: JWT_SECRET_KEY not set — using insecure default")
		key = "change-me-in-production"
	}
	return []byte(key)
}

// CreateToken mints a JWT carrying `sub` (username) and `role`.
func CreateToken(username, role string) (string, error) {
	expiry := config.GetJwtConfig().ExpiryHours
	if expiry <= 0 {
		expiry = 24
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  username,
		"role": role,
		"exp":  time.Now().Add(time.Duration(expiry) * time.Hour).Unix(),
	})

	tokenString, err := claims.SignedString(getSecretKey())
	if err != nil {
		logger.Logger.WithField("user", username).Errorf("token: sign JWT: %v", err)
		return "", err
	}

	logger.Logger.WithField("user", username).Debugf("token: created (expires in %dh)", expiry)
	return prefixKey + tokenString, nil
}

// ParseToken parses a JWT (with or without the "Basic " prefix) and returns
// the username and role from the claims.
func ParseToken(tokenString string) (username string, role string, err error) {
	tokenString = strings.TrimPrefix(tokenString, prefixKey)

	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getSecretKey(), nil
	})
	if err != nil {
		logger.Logger.Errorf("token: parse JWT: %v", err)
		return "", "", err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return "", "", errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", "", errors.New("invalid token: missing sub")
	}
	r, _ := claims["role"].(string)
	if r == "" {
		r = "user"
	}
	return sub, r, nil
}
