package utils

import (
	"context"
	"fmt"

	go_jwt "github.com/golang-jwt/jwt"
)

type JwtPAYLOAD struct {
	Sub   string
	Name  string
	Email string
	Iat   float64
}

type JwtItem struct {
	Token   *go_jwt.Token
	Payload JwtPAYLOAD
}

var hmacSampleSecret []byte

func ParseJWT(tokenString string) (JwtItem, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := go_jwt.Parse(tokenString, func(token *go_jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*go_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if err != nil {
		fmt.Print(err)
		return JwtItem{}, err
	}

	t := token.Claims.(go_jwt.MapClaims)

	h := JwtItem{
		Token: token,
		Payload: JwtPAYLOAD{
			Sub:   t["sub"].(string),
			Name:  t["name"].(string),
			Email: t["email"].(string),
			Iat:   t["iat"].(float64),
		},
	}

	return h, nil
}

func RetrievJWTFromContext(ctx context.Context) (JwtItem, error) {
	svcUntyped := ctx.Value(Jwt)
	jwt, b := svcUntyped.(JwtItem)
	if b {
		return jwt, nil
	} else {
		return JwtItem{}, fmt.Errorf("no valid jwt")
	}

}
