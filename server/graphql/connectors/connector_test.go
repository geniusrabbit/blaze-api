package connectors

import (
	"context"
	"testing"

	"github.com/geniusrabbit/blaze-api/repository"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

func TestPageInfoWithInvalidSize(t *testing.T) {
	t.Parallel()

	accessor := &DataAccessorFunc[int]{
		FetchDataListFunc: func(context.Context) ([]int, error) { return nil, nil },
		CountDataFunc:     func(context.Context) (int64, error) { return 25, nil },
	}
	// Before the fix size 0 (or negative) panicked with a division by zero
	for _, size := range []int{0, -3} {
		info := NewCollectionConnection[int](context.Background(), accessor, &gqlmodels.Page{Size: &size}).PageInfo()
		want := 25/repository.DefaultPageSize + 1
		if info.Count != want {
			t.Errorf("size %d: page count = %d, want %d", size, info.Count, want)
		}
	}
}

func TestNilPageIsCapped(t *testing.T) {
	t.Parallel()

	var page *gqlmodels.Page
	if p := page.Pagination(); p == nil || p.Size != repository.MaxPageSize {
		t.Fatalf("pagination = %+v, want size %d", p, repository.MaxPageSize)
	}
}
