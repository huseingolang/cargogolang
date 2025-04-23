package Hook

import (
	"time"

	"github.com/golang-jwt/jwt"
)

var SecretKay = []byte("dffdgrrr11414")

func GenerateJWT(u string) string {
	createToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokens, err := createToken.SignedString([]byte(SecretKay))
	if err != nil {
		panic(err)

	}

	return tokens
}
