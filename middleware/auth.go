package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var Secret = []byte("your-secret-key")

type Claims struct {
	Email string
	Role string `json:"role"`
	ID uint
	jwt.StandardClaims
}

func CreateToken(role string, email string, id uint)(string, error){
	claims := Claims{
		Email: email,
		Role: role,
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

func AuthMiddlerware(requiredRole string) gin.HandlerFunc{
	return func(c *gin.Context){
		//authHeader := c.GetHeader("Authorization")
		session := sessions.Default(c)
		username := session.Get("name")

		token, err := c.Cookie("JWT")

		if username == nil || err != nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"Authorization header required"})
			c.Abort()
			return 
		}

		
		if token == ""{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Token missing"})
			c.Abort()
			return 
		}

		tokenres, err := jwt.ParseWithClaims(token, &Claims{}, func (token *jwt.Token)(interface{}, error){
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
				return nil, fmt.Errorf("unexpected signing method: %v",token.Header["alg"])
			}
			return Secret, nil
		})

		if err != nil || !tokenres.Valid{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid or expired token"})
			c.Abort()
			return 
		}

		if claims, ok := tokenres.Claims.(*Claims); ok && tokenres.Valid{
			if claims.Role != requiredRole {
				c.JSON(http.StatusForbidden, gin.H{"message":"Insufficient privileges"})
				c.Abort()
				return 
			}
			c.Set("claims",claims)
		}else{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid token claims"})
			c.Abort()
			return 
		}
		c.Next()
	}
}

