package util

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
	_ "time/tzdata"
)

type Jwt interface {
	GenerateAccessToken(userID string, role string) (string, error)
	ValidateAccessToken(token string) (jwt.MapClaims, error)
	GenerateRefreshToken(userID string, role string) (string, error)
	ValidateRefreshToken(token string) (jwt.MapClaims, error)
}

const TOKEN_EXPIRED_TIME = 30 * time.Minute

type JwtImpl struct {
}

func NewJwtImpl() *JwtImpl {
	return &JwtImpl{}
}

func (j JwtImpl) GenerateAccessToken(userID string, role string) (string, error) {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return "", err
	}
	expireTime := time.Now().In(location).Local().Add(1 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    expireTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j JwtImpl) ValidateAccessToken(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("ACCESS_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (j JwtImpl) GenerateRefreshToken(userID string, role string) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    expireTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("REFRESH_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j JwtImpl) ValidateRefreshToken(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("REFRESH_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
