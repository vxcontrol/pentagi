package telegram

type Auth struct {
	allowed map[int64]bool
}

func NewAuth(ids []int64) *Auth {
	m := make(map[int64]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return &Auth{allowed: m}
}

func (a *Auth) IsAllowed(userID int64) bool {
	return a.allowed[userID]
}