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
		return c.secret, nil
	})
	if err != nil {
		return errors.Wrap(err, "Error parsing JWT")
	}
	if !parsedToken.Valid {
		return errors.New("Invalid token")
	}
	expirationToken := parsedToken.Claims.(jwt.MapClaims)["exp"].(float64)
	if int64(expirationToken) < time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}
