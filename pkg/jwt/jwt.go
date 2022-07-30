package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	b64 "encoding/base64"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/tekenradar/researcher-backend/internal/config"
)

var (
	secretKey    []byte
	secretKeyEnc string
)

// UserClaims - Information a token enocodes
type UserClaims struct {
	ID      string   `json:"id,omitempty"`
	Studies []string `json:"studies,omitempty"`
	jwt.StandardClaims
}

func getSecretKey() (newSecretKey []byte, err error) {
	newSecretKeyEnc := os.Getenv(config.ENV_JWT_TOKEN_KEY)
	if secretKeyEnc == newSecretKeyEnc {
		return newSecretKey, nil
	}
	secretKeyEnc = newSecretKeyEnc
	newSecretKey, err = b64.StdEncoding.DecodeString(newSecretKeyEnc)
	if err != nil {
		return newSecretKey, err
	}
	if len(newSecretKey) < 32 {
		return newSecretKey, errors.New("couldn't find proper secret key")
	}
	secretKey = newSecretKey
	return
}

// GenerateNewToken create and signes a new token
func GenerateNewToken(userID string, experiresIn time.Duration, studies []string) (string, error) {
	// Create the Claims
	claims := UserClaims{
		userID,
		studies,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(experiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	_, err := getSecretKey()
	if err != nil {
		return "", err
	}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secretKey)
	return tokenString, err
}

// ValidateToken parses and validates the token string
func ValidateToken(tokenString string) (claims *UserClaims, valid bool, err error) {
	_, err = getSecretKey()
	if err != nil {
		return nil, false, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if token == nil {
		return
	}
	claims, valid = token.Claims.(*UserClaims)
	valid = valid && token.Valid
	return
}
