package auth

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"

	authapi "imanager/pkg/api/auth"
)

var mySigningKey = []byte("token.secret")

func CreateToken(info authapi.RespToken) (tokenss string, err error) {
	//自定义claim
	claim := jwt.MapClaims{
		"info": info,
		"nbf":  info.IssuedAt.Unix(),
		"iat":  info.IssuedAt.Unix(),
		"exp":  info.ExpiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenss, err = token.SignedString(mySigningKey)
	return
}

func ParseToken(tokenss string) (authapi.RespToken, error) {
	if len(tokenss) == 0 {
		return authapi.RespToken{}, fmt.Errorf("token is empty")
	}
	token, err := jwt.Parse(tokenss, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return authapi.RespToken{}, err
	}
	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("cannot convert claim to mapclaim")
		return authapi.RespToken{}, err
	}
	//验证token，如果token被修改过则为false
	if !token.Valid {
		err = errors.New("token is invalid")
		return authapi.RespToken{}, err
	}

	body, err := json.Marshal(claim["info"])
	if err != nil {
		return authapi.RespToken{}, err
	}

	res := authapi.RespToken{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
