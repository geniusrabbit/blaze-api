package directives

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/geniusrabbit/blaze-api/pkg/cache/memory"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
	"github.com/geniusrabbit/blaze-api/server/graphql/context/gqlapicache"
)

type cacheTestAccount struct {
	accountModels.AccountBase
}

func (a *cacheTestAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &cacheTestAccount{}
	acc.ID = id
	acc.Admins = adminUserIDs
	return acc
}

// encodedList implements both cache codec interfaces.
type encodedList struct {
	Items []string
}

func (l *encodedList) EncodableCacheValue() ([]byte, error) { return json.Marshal(l.Items) }
func (l *encodedList) DecodableCacheValue(data []byte) error {
	return json.Unmarshal(data, &l.Items)
}

func newCacheCtx(t *testing.T, cacheCli *memory.Cache, userID, accountID uint64, queryDirective bool) context.Context {
	t.Helper()
	ctx := context.Background()
	ctx = permissions.WithManager(ctx, permissions.NewTestManager(ctx))
	acc := &cacheTestAccount{}
	acc.ID = accountID
	ctx = session.WithUserAccount(ctx, testutil.Stub(userID), acc)
	ctx = gqlapicache.WithGqlApiCache(ctx, cacheCli)
	ctx = graphql.WithOperationContext(ctx, &graphql.OperationContext{OperationName: "Q"})

	field := &ast.Field{Name: "list", Alias: "list"}
	if queryDirective {
		field.Directives = ast.DirectiveList{{Name: "cacheData"}}
	}
	return graphql.WithFieldContext(ctx, &graphql.FieldContext{
		Object: "Query",
		Field:  graphql.CollectedField{Field: field},
		Args:   map[string]any{"id": 1},
	})
}

func newTestCache(t *testing.T) *memory.Cache {
	t.Helper()
	cacheCli, err := memory.NewTimeout(context.Background(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return cacheCli
}

// resolve runs the directive with a resolver returning value and reports
// whether the resolver was called.
func resolve(t *testing.T, ctx context.Context, value any) (any, bool) {
	t.Helper()
	called := false
	res, err := CacheData(ctx, nil, func(context.Context) (any, error) {
		called = true
		return value, nil
	}, 60, nil, nil)
	if err != nil {
		t.Fatalf("CacheData: %v", err)
	}
	return res, called
}

func TestCacheDataIsolatesSessions(t *testing.T) {
	t.Parallel()

	cacheCli := newTestCache(t)

	res, called := resolve(t, newCacheCtx(t, cacheCli, 1, 10, false), "user 1 data")
	if !called || res != "user 1 data" {
		t.Fatalf("first call: res = %v, called = %v", res, called)
	}
	if res, called = resolve(t, newCacheCtx(t, cacheCli, 1, 10, false), "fresh"); called || res != "user 1 data" {
		t.Fatalf("same session: res = %v, called = %v; want a cache hit", res, called)
	}

	// Before the fix the key had no user or account: user 2 got user 1's data
	if res, called = resolve(t, newCacheCtx(t, cacheCli, 2, 10, false), "user 2 data"); !called || res != "user 2 data" {
		t.Fatalf("other user: res = %v, called = %v", res, called)
	}
	if res, called = resolve(t, newCacheCtx(t, cacheCli, 1, 11, false), "account 11 data"); !called || res != "account 11 data" {
		t.Fatalf("other account: res = %v, called = %v", res, called)
	}
}

func TestCacheDataIgnoresQueryDirective(t *testing.T) {
	t.Parallel()

	cacheCli := newTestCache(t)
	for range 2 {
		if _, called := resolve(t, newCacheCtx(t, cacheCli, 1, 10, true), "data"); !called {
			t.Fatal("a @cacheData set by the client was served from the cache")
		}
	}
	// Nothing was stored for the schema-level field either
	if _, called := resolve(t, newCacheCtx(t, cacheCli, 1, 10, false), "data"); !called {
		t.Fatal("a value cached by a client directive was served")
	}
}

func TestCacheDataDecodesEncodableValues(t *testing.T) {
	t.Parallel()

	cacheCli := newTestCache(t)
	ctx := newCacheCtx(t, cacheCli, 1, 10, false)
	if _, called := resolve(t, ctx, &encodedList{Items: []string{"a", "b"}}); !called {
		t.Fatal("first call wasn't resolved")
	}

	// Before the fix CacheDecodable values were never decoded, so every call missed
	res, called := resolve(t, ctx, &encodedList{Items: []string{"fresh"}})
	list, ok := res.(*encodedList)
	if called || !ok || len(list.Items) != 2 || list.Items[1] != "b" {
		t.Fatalf("res = %#v, called = %v; want the cached list", res, called)
	}
}
