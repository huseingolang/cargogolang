package Middleware

import (
	model "chat/Model"
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var SecretKay = []byte("dffdgrrr11414")

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"id"`
}

var UserCtx = "UserId"

func AuthUSer(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	if header == " " {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	headerSplit := strings.Split(header, " ")
	if len(headerSplit) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return

	}
	headerParse, err := ParseJwt(headerSplit[1])
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.Set(UserCtx, headerParse)
}
func ParseJwt(assetoken string) (int, error) {
	token, err := jwt.ParseWithClaims(assetoken, &model.User{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid sign Method")
		}
		return []byte(SecretKay), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, err
	}
	return claims.UserId, nil

}
