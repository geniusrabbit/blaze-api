package middleware

// import (
// 	"context"
// 	"testing"

// 	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
// 	"github.com/stretchr/testify/assert"
// )

// func init() {
// 	SetTestMode(true)
// }

// func Test_AuthContext(t *testing.T) {
// 	ctx := WithGRPCAuthorizationToken(context.TODO(), "test-token")
// 	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test-token", token)
// }
