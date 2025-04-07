package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

type JWTClient interface {
	NewJWT(user CreateTokenParams) (token string, err error)
	ValidateJWT(token string) error
}

type jwtClient struct {
	secret     string
	accessTime time.Duration
}

func NewJwtClient(secret string, accessTime time.Duration) *jwtClient {
	return &jwtClient{secret: secret, accessTime: accessTime}
}

type CreateTokenParams struct {
	Username string
	Role     string
}

func (c *jwtClient) NewJWT(user CreateTokenParams) (token string, err error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(c.accessTime).Unix(),
	})
	token, err = claims.SignedString([]byte(c.secret))
	if err != nil {
		return "", errors.Wrap(err, "Error creating JWT token")
	}
	return token, nil
}

func (c *jwtClient) ValidateJWT(token string) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.secret), nil
	})
	if err != nil {
		return errors.Wrap(err, "Error parsing JWT")
	}
	if !parsedToken.Valid {
		return errors.New("Invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("Invalid claims in token")
	}
	expirationToken, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("Invalid expiration time in token")
	}
	if int64(expirationToken) < time.Now().Unix() {
		return errors.New("token is expired")
	}
	return nil
}
