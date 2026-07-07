package directives

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
)

func ValidateRegex(ctx context.Context, obj any, next graphql.Resolver, pattern string, trim, ornil bool) (res any, err error) {
	if res, err = next(ctx); err != nil {
		return nil, err
	}

	var (
		str      string
		canBeNil = false
	)

	// Check if the value is a string or a pointer to a string
	switch v := res.(type) {
	case nil:
		if ornil {
			return nil, nil
		}
		return nil, ErrValueIsNil
	case string:
		str = v
	case *string:
		str = *v
		canBeNil = true
	default:
		if gocast.IsEmpty(res) {
			if ornil && reflect.ValueOf(res).Kind() == reflect.Ptr {
				return nil, nil
			}
			return nil, ErrValueIsEmpty
		}
	}

	// Check if the string is empty
	if str == "" && ornil && canBeNil {
		return nil, nil
	}

	if !regexp.MustCompile(pattern).MatchString(str) {
		return nil, fmt.Errorf("value does not match pattern `%s`", pattern)
	}

	if canBeNil {
		return &str, nil
	}
	return str, nil
}
