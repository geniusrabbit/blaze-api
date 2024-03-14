package serverprovider

import (
	"context"

	"github.com/ory/fosite"

	"github.com/geniusrabbit/blaze-api/model"
)

var ctxTarget = struct{ s string }{"oauth2:user"}

type target struct {
	uid        uint64
	clientObj  *model.AuthClient
	sessionObj *model.AuthSession
}

// NewContext with additional functionality for oauth2 module
func NewContext(ctxs ...context.Context) context.Context {
	var ctx context.Context
	if len(ctxs) > 0 && ctxs[0] != nil {
		ctx = ctxs[0]
	} else {
		ctx = fosite.NewContext()
	}
	ctx = context.WithValue(ctx, ctxTarget, &target{})
	return ctx
}

// SetContextTargetUserID puts user ID into the context to reuse it in future
func SetContextTargetUserID(ctx context.Context, id uint64) {
	getContextTarget(ctx).uid = id
}

// GetContextTargetUserID returns ID of the user
func GetContextTargetUserID(ctx context.Context) uint64 {
	return getContextTarget(ctx).uid
}

// SetContextTargetClient puts user ID into the context to reuse it in future
func SetContextTargetClient(ctx context.Context, client *model.AuthClient) {
	getContextTarget(ctx).clientObj = client
}

// GetContextTargetClient returns auth client object
func GetContextTargetClient(ctx context.Context) *model.AuthClient {
	return getContextTarget(ctx).clientObj
}

// SetContextSession puts session model into the context
func SetContextSession(ctx context.Context, session *model.AuthSession) {
	getContextTarget(ctx).sessionObj = session
}

// GetContextSession returns session model object from context
func GetContextSession(ctx context.Context) *model.AuthSession {
	return getContextTarget(ctx).sessionObj
}

func getContextTarget(ctx context.Context) *target {
	tr := ctx.Value(ctxTarget)
	switch v := tr.(type) {
	case *target:
		return v
	default:
		panic(`invalid oauth2 context, please use "oauth2.NewContext()"`)
	}
}
