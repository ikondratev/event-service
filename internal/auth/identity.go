package auth

import "context"

type Identity struct {
	Subject string
	Email 	string
	Scopes 	[]string
}

type ctxKey struct {}

func (i Identity) HasScope(want string) bool {
	for _, s := range i.Scopes {
		if want == s {
			return true
		}
	}
	return false
}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}
