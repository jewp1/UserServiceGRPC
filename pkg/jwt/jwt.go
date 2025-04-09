package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

type JWTClient interface {
	NewJWT(user CreateTokenParams) (tokens *CreateTokenResponse, err error)
	ValidateJWT(token string) error
}

type jwtClient struct {
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

func NewJwtClient(secret string, refreshTime time.Duration, accessTime time.Duration) *jwtClient {
	return &jwtClient{secret: secret, refreshTime: refreshTime, accessTime: accessTime}
}

type CreateTokenParams struct {
	Username string
	Role     string
}

type CreateTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

func (c *jwtClient) NewJWT(user CreateTokenParams) (*CreateTokenResponse, error) {
	accessToken, err := c.newToken(user, c.accessTime)
	if err != nil {
		return nil, errors.New("Error creating access token" + err.Error())
	}
	refreshToken, err := c.newToken(user, c.refreshTime)
	if err != nil {
		return nil, errors.New("Error creating refresh token" + err.Error())
	}
	return &CreateTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
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
	expirationToken := parsedToken.Claims.(jwt.MapClaims)["exp"].(float64)
	if int64(expirationToken) < time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}

func (c *jwtClient) newToken(user CreateTokenParams, It time.Duration) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(It).Unix(),
	})
	token, err := claims.SignedString([]byte(c.secret))
	if err != nil {
		return "", errors.Wrap(err, "Error signing JWT token")
	}
	return token, nil
}
