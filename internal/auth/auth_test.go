package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeJWT(t *testing.T) {
	// Test data
	userID := uuid.New()
	secret := "secretkey"
	expiresIn := 1 * time.Hour

	// Create JWT
	token, err := MakeJWT(userID, secret, expiresIn)
	require.NoError(t, err, "Error creating JWT")

	// Parse and validate token
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	// Check if the parsed token is valid and contains correct userID
	assert.NoError(t, err, "Error parsing JWT")
	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	assert.True(t, ok, "Token claims should be of type *jwt.RegisteredClaims")
	assert.Equal(t, userID.String(), claims.Subject, "User ID does not match")
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	// Test data
	userID := uuid.New()
	secret := "secretkey"
	expiresIn := -time.Hour // Token that expires in the past

	// Create JWT
	token, err := MakeJWT(userID, secret, expiresIn)
	require.NoError(t, err, "Error creating JWT")

	// Validate token
	_, err = ValidateJWT(token, secret)
	assert.Error(t, err, "Expected error for expired token")
	assert.Equal(t, err.Error(), "token has invalid claims: token is expired", "Error message should indicate invalid token")
}

func TestValidateJWT_InvalidSecret(t *testing.T) {
	// Test data
	userID := uuid.New()
	secret := "secretkey"
	wrongSecret := "wrongsecret"
	expiresIn := 1 * time.Hour

	// Create JWT
	token, err := MakeJWT(userID, secret, expiresIn)
	require.NoError(t, err, "Error creating JWT")

	// Validate token with wrong secret
	_, err = ValidateJWT(token, wrongSecret)
	assert.Error(t, err, "Expected error for invalid secret")
	assert.Equal(t, err.Error(), "token signature is invalid: signature is invalid", "Error message should indicate invalid token")
}

func TestValidateJWT_ValidToken(t *testing.T) {
	// Test data
	userID := uuid.New()
	secret := "secretkey"
	expiresIn := 1 * time.Hour

	// Create JWT
	token, err := MakeJWT(userID, secret, expiresIn)
	require.NoError(t, err, "Error creating JWT")

	// Validate token with correct secret
	parsedUserID, err := ValidateJWT(token, secret)
	require.NoError(t, err, "Expected valid token")
	assert.Equal(t, userID, parsedUserID, "User ID does not match")
}
