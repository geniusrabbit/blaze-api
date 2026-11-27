package limits

import (
	"context"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func opContext(t *testing.T, query string) *graphql.OperationContext {
	t.Helper()
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		t.Fatal(err)
	}
	return &graphql.OperationContext{Doc: doc, Operation: doc.Operations[0]}
}

func TestShapeLimit(t *testing.T) {
	t.Parallel()

	limit := &ShapeLimit{MaxDepth: 3, MaxAliases: 2}
	cases := map[string]struct {
		query string
		err   string
	}{
		"flat":             {query: `{ a b c }`},
		"depth at limit":   {query: `{ a { b { c } } }`},
		"too deep":         {query: `{ a { b { c { d } } } }`, err: "depth 4"},
		"deep fragment":    {query: `query { a { ...F } } fragment F on T { b { c { d } } }`, err: "depth 4"},
		"inline fragment":  {query: `{ a { ... on T { b { c { d } } } } }`, err: "depth 4"},
		"aliases at limit": {query: `{ x: a y: a }`},
		"too many aliases": {query: `{ x: a y: a z: a }`, err: "3 aliases"},
		"fragment aliases": {query: `query { ...F ...F } fragment F on Q { x: a y: a }`, err: "4 aliases"},
		"fragment cycle":   {query: `query { ...F } fragment F on Q { a ...F }`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gqlErr := limit.MutateOperationContext(context.Background(), opContext(t, tc.query))
			switch {
			case tc.err == "" && gqlErr != nil:
				t.Fatalf("unexpected error: %v", gqlErr)
			case tc.err != "" && (gqlErr == nil || !strings.Contains(gqlErr.Message, tc.err)):
				t.Fatalf("error = %v, want %q", gqlErr, tc.err)
			}
		})
	}
}

func TestShapeLimitDisabled(t *testing.T) {
	t.Parallel()

	limit := &ShapeLimit{}
	if err := limit.MutateOperationContext(context.Background(), opContext(t, `{ a { b { c { d { e x: f y: f z: f } } } } }`)); err != nil {
		t.Fatalf("disabled limits returned %v", err)
	}
}

func TestConfigLimit(t *testing.T) {
	t.Parallel()

	cfg := Config{MaxDepth: -1, MaxAliases: 5}
	if got := cfg.limit(cfg.MaxDepth, DefaultMaxDepth); got != -1 {
		t.Errorf("negative limit = %d, want -1 (off)", got)
	}
	if got := cfg.limit(cfg.MaxComplexity, DefaultMaxComplexity); got != DefaultMaxComplexity {
		t.Errorf("zero limit = %d, want the default", got)
	}
	if got := cfg.limit(cfg.MaxAliases, DefaultMaxAliases); got != 5 {
		t.Errorf("set limit = %d, want 5", got)
	}
}
