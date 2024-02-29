package models

import "github.com/dgrijalva/jwt-go"

type JWTToken struct{
	Username string
	jwt.StandardClaims
}
