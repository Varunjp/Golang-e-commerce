package middleware

import (
	"time"

	"github.com/golang-jwt/jwt"
)

var Secret = []byte("your-secret-key")

type Claims struct {
	Email string
	ID uint
	jwt.StandardClaims
}

func CreateToken(email string, id uint)(string, error){
	claims := Claims{
		Email: email,
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt: time.Now().Unix(),
			Issuer: "Fashion Art",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return token.SignedString(Secret)
}