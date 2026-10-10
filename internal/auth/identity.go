package auth

type Identity struct {
	Subject string
	Email   string
	Scopes  []string
}

func (i Identity) HasScope(want string) bool {
	for _, s := range i.Scopes {
		if want == s {
			return true
		}
	}
	return false
}
