package services

import (
	models "backend/Models"
	"context"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var secretKey = []byte("SecretKey") //TODO Secret Key Will unique

type TokenService interface {
	GenerateToken(userName string) (string, error)
	AuthenticateJWT() gin.HandlerFunc
}

type TokenServiceImplementation struct {
	ctx context.Context
}

func NewTokenService(ctx context.Context,) TokenService {
	return &TokenServiceImplementation{
		ctx: ctx,
	}
}

func (TokenServiceImplementation) GenerateToken(userName string) (string, error) {
	expirationTime := time.Now().Add(time.Hour)
	wt := &models.JWTToken{
		Username: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, wt)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (ts *TokenServiceImplementation) AuthenticateJWT() gin.HandlerFunc {
	return func (c *gin.Context) {
		const BearerSchema = "Bearer "
		header := c.GetHeader("Authorization");
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}
		tokenString := header[len(BearerSchema):]
		jwtToken := &models.JWTToken{}

		token, err := jwt.ParseWithClaims(tokenString, jwtToken, func(token *jwt.Token) (interface {}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("username", jwtToken.Username)
		c.Next();
	}
}
