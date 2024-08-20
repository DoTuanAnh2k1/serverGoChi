package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"serverGoChi/src/log"
	"strings"
	"time"
)

var secretKey = []byte("optimus-prime-auto-bot")
var prefixKey = "Basic "

func CreateToken(username string, roles string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,                         // Subject (user identifier)
		"aud": roles,                            // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
	})

	// Print information about the created token
	log.Logger.Info("Token claims added: %+v\n", claims)

	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		log.Logger.Error("Cannot create json web token: ", err)
		return "", err
	}

	tokenString = prefixKey + tokenString

	return tokenString, nil
}

// ParseToken parses the JWT token and returns the username and roles
func ParseToken(tokenString string) (string, string, error) {
	// Remove the "Basic " prefix if it exists
	if strings.HasPrefix(tokenString, prefixKey) {
		tokenString = strings.TrimPrefix(tokenString, prefixKey)
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Logger.Error("unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		log.Logger.Error("Error parsing token: ", err)
		return "", "", err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Retrieve the username and roles from the claims
		username := claims["sub"].(string)
		roles := claims["aud"].(string)
		return username, roles, nil
	}

	return "", "", errors.New("invalid token")
}
