package jwt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewJWT(t *testing.T) {
	client := NewJwtClient("test_secret", 1*time.Hour, 1*time.Hour)

	tokens, err := client.NewJWT(CreateTokenParams{
		Username: "testuser",
		Role:     "admin",
	})
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
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
				return token + "corrupted"
			},
			expectErr:   true,
			expectedErr: "Error parsing JWT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewJwtClient("test_secret", 1*time.Hour, tc.accessTime)

			tokens, err := client.NewJWT(CreateTokenParams{
				Username: "testuser",
				Role:     "admin",
			})
			assert.NoError(t, err)

			tokenToUse := tokens.AccessToken
			if tc.modifyToken != nil {
				tokenToUse = tc.modifyToken(tokenToUse)
			}

			err = client.ValidateJWT(tokenToUse)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
