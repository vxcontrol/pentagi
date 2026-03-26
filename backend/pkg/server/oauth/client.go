package oauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

type OAuthEmailResolver func(ctx context.Context, nonce string, token *oauth2.Token) (string, error)

type OAuthClient interface {
	ProviderName() string
	ResolveEmail(ctx context.Context, nonce string, token *oauth2.Token) (string, error)
	TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	RefreshToken(ctx context.Context, token string) (*oauth2.Token, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
}

type oauthClient struct {
	name          string
	verifier      string
	conf          *oauth2.Config
	emailResolver OAuthEmailResolver
}

func NewOAuthClient(name string, conf *oauth2.Config, emailResolver OAuthEmailResolver) OAuthClient {
	return &oauthClient{
		name:          name,
		verifier:      oauth2.GenerateVerifier(),
		conf:          conf,
		emailResolver: emailResolver,
	}
}

func (o *oauthClient) ProviderName() string {
	return o.name
}

func (o *oauthClient) ResolveEmail(ctx context.Context, nonce string, token *oauth2.Token) (string, error) {
	if o.emailResolver == nil {
		return "", fmt.Errorf("email resolver is not set")
	}
	return o.emailResolver(ctx, nonce, token)
}

func (o *oauthClient) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return o.conf.TokenSource(ctx, token)
}

func (o *oauthClient) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	opts = append(opts, oauth2.VerifierOption(o.verifier))
	return o.conf.Exchange(ctx, code, opts...)
}

func (o *oauthClient) RefreshToken(ctx context.Context, token string) (*oauth2.Token, error) {
	return o.conf.TokenSource(ctx, &oauth2.Token{RefreshToken: token}).Token()
}

func (o *oauthClient) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	opts = append(opts, oauth2.S256ChallengeOption(o.verifier))
	return o.conf.AuthCodeURL(state, opts...)
}
