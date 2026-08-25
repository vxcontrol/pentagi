package oauth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type OAuthEmailResolver func(ctx context.Context, nonce string, token *oauth2.Token) (string, bool, error)

type OAuthClient interface {
	ProviderName() string
	ResolveEmail(ctx context.Context, nonce string, token *oauth2.Token) (string, bool, error)
	TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	RefreshToken(ctx context.Context, token string) (*oauth2.Token, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	// AuthCodeOptions returns provider specific parameters for the authorization
	// request, such as the nonce or the response mode the provider expects.
	AuthCodeOptions(nonce string) []oauth2.AuthCodeOption
	// CallbackSameSite reports which SameSite mode the temporary state and nonce
	// cookies need so that they survive the provider's callback. Providers that
	// answer with a cross-site form POST require SameSite=None.
	CallbackSameSite() http.SameSite
}

// ClientOption customizes provider specific behaviour of an OAuth client.
type ClientOption func(*oauthClient)

// WithFormPostCallback marks the provider as answering with a cross-site form
// POST (OpenID Connect `response_mode=form_post`) and requests an ID token
// alongside the authorization code. Such a callback only carries the state and
// nonce cookies when they are issued with SameSite=None.
func WithFormPostCallback() ClientOption {
	return func(o *oauthClient) {
		o.formPostCallback = true
	}
}

type oauthClient struct {
	name             string
	verifier         string
	conf             *oauth2.Config
	emailResolver    OAuthEmailResolver
	formPostCallback bool
}

func NewOAuthClient(name string, conf *oauth2.Config, emailResolver OAuthEmailResolver, opts ...ClientOption) OAuthClient {
	client := &oauthClient{
		name:          name,
		verifier:      oauth2.GenerateVerifier(),
		conf:          conf,
		emailResolver: emailResolver,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (o *oauthClient) ProviderName() string {
	return o.name
}

func (o *oauthClient) ResolveEmail(ctx context.Context, nonce string, token *oauth2.Token) (string, bool, error) {
	if o.emailResolver == nil {
		return "", false, fmt.Errorf("email resolver is not set")
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

func (o *oauthClient) AuthCodeOptions(nonce string) []oauth2.AuthCodeOption {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", nonce),
	}
	if o.formPostCallback {
		opts = append(opts,
			oauth2.SetAuthURLParam("response_mode", "form_post"),
			oauth2.SetAuthURLParam("response_type", "code id_token"),
		)
	}
	return opts
}

func (o *oauthClient) CallbackSameSite() http.SameSite {
	if o.formPostCallback {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}
