package jwt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewJWT(t *testing.T) {
	client := NewJwtClient("test_secret", 1*time.Hour)

	token, err := client.NewJWT(CreateTokenParams{
		Username: "testuser",
		Role:     "admin",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateJWT(t *testing.T) {
	testCases := []struct {
		name        string
		accessTime  time.Duration
		modifyToken func(string) string
		expectErr   bool
		expectedErr string
	}{
		{
			name:        "Valid token",
			accessTime:  1 * time.Hour,
			modifyToken: nil,
			expectErr:   false,
		},
		{
			name:        "Expired token",
			accessTime:  -1 * time.Second,
			modifyToken: nil,
			expectErr:   true,
			expectedErr: "token is expired",
		},
		{
			name:       "Invalid token format",
			accessTime: 1 * time.Hour,
			modifyToken: func(token string) string {
				return token + "invalid token"
			},
			expectErr:   true,
			expectedErr: "Error parsing JWT",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			client := NewJwtClient("test_secret", test.accessTime)

			token, err := client.NewJWT(CreateTokenParams{
				Username: "testuser",
				Role:     "admin",
			})
			assert.NoError(t, err)

			if test.modifyToken != nil {
				token = test.modifyToken(token)
			}

			err = client.ValidateJWT(token)

			if test.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
