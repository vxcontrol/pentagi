package oauth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProviderName is the provider name reported to the frontend for a generic
// OpenID Connect identity provider such as Keycloak, Authentik, Okta or Entra ID.
const OIDCProviderName = "oidc"

// DefaultOIDCScopes are requested when no explicit scope list is configured.
// The email scope is what makes the email claim available, and PentAGI matches
// users by email.
var DefaultOIDCScopes = []string{oidc.ScopeOpenID, "email", "profile"}

type oidcTokenClaims struct {
	Nonce         string `json:"nonce"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func newOIDCEmailResolver(provider *oidc.Provider, clientID string) OAuthEmailResolver {
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return func(ctx context.Context, nonce string, token *oauth2.Token) (string, bool, error) {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			return "", false, fmt.Errorf("id_token is not present in the token")
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return "", false, fmt.Errorf("could not verify OIDC ID Token: %w", err)
		}

		if idToken.Nonce != nonce {
			return "", false, fmt.Errorf("nonce mismatch in OIDC ID Token")
		}

		claims := oidcTokenClaims{}
		if err := idToken.Claims(&claims); err != nil {
			return "", false, fmt.Errorf("failed to parse OIDC ID Token claims: %w", err)
		}

		if claims.Nonce != nonce {
			return "", false, fmt.Errorf("nonce mismatch in OIDC ID Token claims")
		}

		if claims.Email != "" {
			return claims.Email, claims.EmailVerified, nil
		}

		// Not every identity provider puts the email into the ID token, so fall
		// back to the UserInfo endpoint before giving up.
		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err != nil {
			return "", false, fmt.Errorf("email is empty in OIDC ID Token claims and UserInfo failed: %w", err)
		}

		if err := userInfo.Claims(&claims); err != nil {
			return "", false, fmt.Errorf("failed to parse OIDC UserInfo claims: %w", err)
		}

		if claims.Email == "" {
			return "", false, fmt.Errorf("email is empty in OIDC ID Token and UserInfo claims")
		}

		return claims.Email, claims.EmailVerified, nil
	}
}

// NewOIDCOAuthClient builds an OAuth client for any OpenID Connect provider by
// reading its discovery document. The issuer is the URL that serves
// /.well-known/openid-configuration, for example
// https://keycloak.example.com/realms/pentagi.
func NewOIDCOAuthClient(
	ctx context.Context,
	issuer, clientID, clientSecret, redirectURL string,
	scopes []string,
) (OAuthClient, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("could not create OpenID client for issuer '%s': %w", issuer, err)
	}

	if len(scopes) == 0 {
		scopes = DefaultOIDCScopes
	}

	return NewOAuthClient(OIDCProviderName, &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     provider.Endpoint(),
	}, newOIDCEmailResolver(provider, clientID)), nil
}
