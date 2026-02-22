package database

type APITokenWithSecret struct {
	ApiToken
	Token string `json:"token"`
}
