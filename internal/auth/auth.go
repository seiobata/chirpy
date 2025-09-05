package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPass), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer")
	if !ok {
		return "", errors.New("token not found")
	}
	return strings.TrimSpace(token), nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	apiKey, ok := strings.CutPrefix(header, "ApiKey")
	if !ok {
		return "", errors.New("api key missing")
	}
	return strings.TrimSpace(apiKey), nil
}
