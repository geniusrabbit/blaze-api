// Package limits protects a gqlgen server against expensive queries:
// query complexity, selection depth and alias count limits, and introspection
// that can be turned off in production.
package limits

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Default limits, suitable for most APIs.
const (
	DefaultMaxComplexity = 1000
	DefaultMaxDepth      = 12
	DefaultMaxAliases    = 30
)

// Config of the query limits. A zero value of a limit means the default;
// a negative value disables that limit.
type Config struct {
	MaxComplexity int
	MaxDepth      int
	MaxAliases    int
	// Introspection enables the schema introspection queries. Keep it off
	// in production unless the schema is public.
	Introspection bool
}

func (c Config) limit(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

// Apply registers the limits on the server.
func Apply(srv *handler.Server, cfg Config) {
	if cfg.Introspection {
		srv.Use(extension.Introspection{})
	}
	if limit := cfg.limit(cfg.MaxComplexity, DefaultMaxComplexity); limit > 0 {
		srv.Use(extension.FixedComplexityLimit(limit))
	}
	depth := cfg.limit(cfg.MaxDepth, DefaultMaxDepth)
	aliases := cfg.limit(cfg.MaxAliases, DefaultMaxAliases)
	if depth > 0 || aliases > 0 {
		srv.Use(&ShapeLimit{MaxDepth: depth, MaxAliases: aliases})
	}
}

// ShapeLimit rejects operations nested deeper than MaxDepth or using more
// than MaxAliases aliases. Fragments are expanded. Non-positive limits are off.
type ShapeLimit struct {
	MaxDepth   int
	MaxAliases int
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationContextMutator
} = (*ShapeLimit)(nil)

// ExtensionName implements graphql.HandlerExtension.
func (*ShapeLimit) ExtensionName() string { return "ShapeLimit" }

// Validate implements graphql.HandlerExtension.
func (*ShapeLimit) Validate(graphql.ExecutableSchema) error { return nil }

// MutateOperationContext checks the operation shape before execution.
func (l *ShapeLimit) MutateOperationContext(_ context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	if rc.Operation == nil {
		return nil
	}
	w := &shapeWalker{doc: rc.Doc}
	depth := w.walk(rc.Operation.SelectionSet, 1, map[string]bool{})
	if l.MaxDepth > 0 && depth > l.MaxDepth {
		return gqlerror.Errorf("operation depth %d exceeds the limit of %d", depth, l.MaxDepth)
	}
	if l.MaxAliases > 0 && w.aliases > l.MaxAliases {
		return gqlerror.Errorf("operation uses %d aliases, the limit is %d", w.aliases, l.MaxAliases)
	}
	return nil
}

type shapeWalker struct {
	doc     *ast.QueryDocument
	aliases int
}

// walk returns the maximum depth of the selection set and counts aliases.
// visiting holds the fragments on the current path to stop on cycles.
func (w *shapeWalker) walk(set ast.SelectionSet, depth int, visiting map[string]bool) int {
	maxDepth := depth - 1
	for _, sel := range set {
		switch s := sel.(type) {
		case *ast.Field:
			if s.Alias != "" && s.Alias != s.Name {
				w.aliases++
			}
			d := depth
			if len(s.SelectionSet) > 0 {
				d = w.walk(s.SelectionSet, depth+1, visiting)
			}
			maxDepth = max(maxDepth, d)
		case *ast.InlineFragment:
			maxDepth = max(maxDepth, w.walk(s.SelectionSet, depth, visiting))
		case *ast.FragmentSpread:
			if visiting[s.Name] || w.doc == nil {
				continue
			}
			frag := w.doc.Fragments.ForName(s.Name)
			if frag == nil {
				continue
			}
			visiting[s.Name] = true
			maxDepth = max(maxDepth, w.walk(frag.SelectionSet, depth, visiting))
			delete(visiting, s.Name)
		}
	}
	return maxDepth
}
