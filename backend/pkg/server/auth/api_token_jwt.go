package auth

import (
	"errors"
	"fmt"
	"time"

	"pentagi/pkg/server/models"

	"github.com/golang-jwt/jwt/v5"
)

func MakeAPIToken(globalSalt string, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(MakeJWTSigningKey(globalSalt))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func MakeAPITokenClaims(tokenID, uhash string, uid, rid, ttl uint64) jwt.Claims {
	now := time.Now()
	return models.APITokenClaims{
		TokenID: tokenID,
		RID:     rid,
		UID:     uid,
		UHASH:   uhash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "api_token",
		},
	}
}

func ValidateAPIToken(tokenString, globalSalt string) (*models.APITokenClaims, error) {
	var claims models.APITokenClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		// verify signing algorithm to prevent "alg: none"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return MakeJWTSigningKey(globalSalt), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token is malformed")
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token is either expired or not active yet")
		} else {
			return nil, fmt.Errorf("token invalid: %w", err)
		}
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return &claims, nil
}
