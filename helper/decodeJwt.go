package helper

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte("your_secret_key")

func DecodeJWT(tokenstr string) (string, float64,error) {

	token, err := jwt.Parse(tokenstr,func(token *jwt.Token)(interface{},error){
		if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret,nil 
	})

	if err != nil || !token.Valid{
		return "",0,errors.New("Error :"+ err.Error())
	}

	claims,ok := token.Claims.(jwt.MapClaims); 
	if !ok{
		return "",0,fmt.Errorf("error while getting claim")
	}

	
	return claims["Email"].(string),claims["ID"].(float64),nil

}