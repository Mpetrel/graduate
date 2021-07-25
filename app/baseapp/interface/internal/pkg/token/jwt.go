// Package token 根据 JSON Web Token 规范生成 jwt token
package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

const secret = "e8b64d06d58a78294b4d9b7c372edb4c"


// Generate 生成token
func Generate(payload map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * time.Duration(240)).Unix(),
	}
	for k, v := range payload {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}


func Verify(tokenStr string) (uid string, valid bool) {
	valid = false
	if tokenStr == "" {
		return
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		log.Printf("verify token failed, \ntoken is: %s\n error: %v\n", tokenStr, err)
		return
	}
	// claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid = claims["id"].(string)
		valid = true
	}
	return
}
