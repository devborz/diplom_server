package utils

import (
	. "clouderrors"
	"fmt"
	"net/mail"
	"os"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func IsExecutable(filePath string) (bool, error) {
    info, err := os.Stat(filePath)
    if err != nil {
        return false, err
    }

    mode := info.Mode()
    return mode&0111 != 0, nil
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

var secretKey = []byte("secretpassword")

// GenerateToken generates a JWT token with the user ID as part of the claims
func GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{}
	claims["user_id"] = userID
    claims["time"] = time.Now().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// VerifyToken verifies a token JWT validate
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}

		return secretKey, nil
	})

	// Check for errors
	if err != nil {
		return nil, err
	}

	// Validate the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePassword(password string) Error {
	if len(password) < 16 {
		return ErrShortPassword
	}
	var hasUpper = false
	var hasLower = false
	var hasSymbol = false
	var hasDigit = false
	rs := []rune(password)
	for i := 0; i < len(rs); i++ {
		var ch = rs[i]
		if unicode.IsUpper(ch) {
			fmt.Println("upper")
			hasUpper = true
		} else if unicode.IsLower(ch) {
			fmt.Println("lower")
			hasLower = true
		} else if unicode.IsDigit(ch) {
			fmt.Println("digit")
			hasDigit = true
		} else {
			hasSymbol = true
        }
	}
	if hasLower && hasUpper && hasSymbol && hasDigit {
		return nil
	} else {
		return ErrWrongPasswordPolicy
	}
}

// func CheckResourceExists(path string) bool {

// }

func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func DeleteResource(path string) error {
	return os.RemoveAll(path)
}
