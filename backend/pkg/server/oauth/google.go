package oauth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleTokenClaims struct {
	Nonce         string `json:"nonce"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func newGoogleEmailResolver(clientID string) OAuthEmailResolver {
	return func(ctx context.Context, nonce string, token *oauth2.Token) (string, bool, error) {
		provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			return "", false, fmt.Errorf("could not create Google OpenID client: %w", err)
		}

		oidToken, ok := token.Extra("id_token").(string)
		if !ok {
			return "", false, fmt.Errorf("id_token is not present in the token")
		}

		verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
		idToken, err := verifier.Verify(ctx, oidToken)
		if err != nil {
			return "", false, fmt.Errorf("could not verify Google ID Token: %w", err)
		}

		if idToken.Nonce != nonce {
			return "", false, fmt.Errorf("nonce mismatch in Google ID Token")
		}

		if err = idToken.VerifyAccessToken(token.AccessToken); err != nil {
			return "", false, fmt.Errorf("failed to verify Google Access Token: %w", err)
		}

		claims := googleTokenClaims{}
		if err := idToken.Claims(&claims); err != nil {
			return "", false, fmt.Errorf("failed to parse Google ID Token claims: %w", err)
		}

		if claims.Nonce != nonce {
			return "", false, fmt.Errorf("nonce mismatch in Google ID Token claims")
		}

		if claims.Email == "" {
			return "", false, fmt.Errorf("email is empty in Google ID Token claims")
		}

		return claims.Email, claims.EmailVerified, nil
	}
}

func NewGoogleOAuthClient(clientID, clientSecret, redirectURL string) OAuthClient {
	return NewOAuthClient("google", &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"openid",
		},
		Endpoint: google.Endpoint,
	}, newGoogleEmailResolver(clientID))
}
