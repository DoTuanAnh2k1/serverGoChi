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

func CreateToken(username string, roles string) (string, error) {
	expiry := config.GetJwtConfig().ExpiryHours
	if expiry <= 0 {
		expiry = 24
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"aud": roles,
		"exp": time.Now().Add(time.Duration(expiry) * time.Hour).Unix(),
	})

	tokenString, err := claims.SignedString(getSecretKey())
	if err != nil {
		logger.Logger.WithField("user", username).Errorf("token: sign JWT: %v", err)
		return "", err
	}

	logger.Logger.WithField("user", username).Debugf("token: created (expires in %dh)", expiry)
	return prefixKey + tokenString, nil
}

// ParseToken parses a JWT (with or without "Basic " prefix) and returns username, roles.
func ParseToken(tokenString string) (string, string, error) {
	tokenString = strings.TrimPrefix(tokenString, prefixKey)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getSecretKey(), nil
	})
	if err != nil {
		logger.Logger.Errorf("token: parse JWT: %v", err)
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token")
	}
	return claims["sub"].(string), claims["aud"].(string), nil
}
