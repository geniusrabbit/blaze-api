package directives

import (
	"context"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
)

func ValidateNotEmpty(ctx context.Context, obj any, next graphql.Resolver, trim, ornil bool) (res any, err error) {
	if res, err = next(ctx); err != nil {
		return nil, err
	}

	// Check if the value is a string or a pointer to a string
	switch res.(type) {
	case nil:
		if ornil {
			return nil, nil
		}
		return nil, ErrValueIsNil
	default:
		if gocast.IsEmpty(res) {
			kind := reflect.ValueOf(res).Kind()
			if ornil && (kind == reflect.Pointer ||
				kind == reflect.Slice ||
				kind == reflect.Map ||
				kind == reflect.Array ||
				kind == reflect.Chan ||
				kind == reflect.Func ||
				kind == reflect.Interface) {
				return nil, nil
			}
			return nil, ErrValueIsEmpty
		}
	}

	return _validateLength(res, 1, 0, trim, ornil)
}
