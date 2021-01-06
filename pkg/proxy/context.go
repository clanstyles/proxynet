package proxy

import "context"

type key int

const userAuthorizationKey key = 0

func FromContext(ctx context.Context) (*Authorization, bool) {
	auth, ok := ctx.Value(userAuthorizationKey).(*Authorization)
	return auth, ok
}

func NewContext(ctx context.Context, auth *Authorization) context.Context {
	return context.WithValue(ctx, userAuthorizationKey, auth)
}
