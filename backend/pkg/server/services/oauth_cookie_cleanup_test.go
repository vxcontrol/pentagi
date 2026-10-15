package services

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuth_OAuthSuccessRemovesTemporaryCookiesFromTheJar(t *testing.T) {
	for _, provider := range []string{"github", "google"} {
		for _, prefix := range []string{"", "tenant_"} {
			t.Run(provider+"/"+prefix, func(t *testing.T) {
				db := setupTestDB(t)
				defer db.Close()
				svc := newOAuthServiceVerified(db, "new@corp.com", true)
				svc.cfg.BaseURL = "/api"
				svc.cfg.LoginCallbackURL = "/auth/login-callback"
				svc.cfg.CookiePrefix = prefix
				svc.oauth[provider] = &fakeOAuthClient{name: provider, email: "new@corp.com", verified: true}
				callbackURL, err := url.Parse("https://pentagi.test/api/auth/login-callback")
				require.NoError(t, err)
				jar, err := cookiejar.New(nil)
				require.NoError(t, err)
				jar.SetCookies(callbackURL, []*http.Cookie{
					{Name: svc.stateCookieName(), Value: "test-state", Path: callbackURL.Path, Secure: true, MaxAge: 300},
					{Name: svc.nonceCookieName(), Value: "test-nonce", Path: callbackURL.Path, Secure: true, MaxAge: 300},
				})
				require.Len(t, jar.Cookies(callbackURL), 2)
				c, rec := newCallbackContext(t)
				c.Request = httptest.NewRequest(http.MethodGet, callbackURL.String(), nil)
				c.Request.Header.Set("X-Forwarded-Proto", "https")
				for _, cookie := range jar.Cookies(callbackURL) {
					c.Request.AddCookie(cookie)
				}
				svc.authLoginCallback(c, map[string]string{"provider": provider}, "test-code")
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				jar.SetCookies(callbackURL, rec.Result().Cookies())
				var remaining []string
				for _, cookie := range jar.Cookies(callbackURL) {
					remaining = append(remaining, cookie.Name)
				}
				require.Equal(t, []string{"pentagi"}, remaining, "only the authenticated session should remain")
			})
		}
	}
}
