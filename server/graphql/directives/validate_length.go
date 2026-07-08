package directives

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
)

func ValidateLength(ctx context.Context, obj any, next graphql.Resolver, min, max int, trim, ornil bool) (res any, err error) {
	if res, err = next(ctx); err != nil {
		return nil, err
	}
	return _validateLength(res, min, max, trim, ornil)
}

func _validateLength(val any, min, max int, trim, ornil bool) (any, error) {
	var (
		res      = val
		str      string
		isArray  = false
		length   int
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
		if isArray = gocast.IsSlice(res); !isArray {
			return nil, fmt.Errorf("value is not a string or slice")
		}
	}

	if isArray {
		length = len(gocast.AnySlice[any](res))
	} else {
		// Trim the string if needed
		if trim {
			str = strings.TrimSpace(str)
		}

		if length = len(str); canBeNil {
			res = &str
		} else {
			res = str
		}
	}

	// Check if the value is empty and can be nil
	if length == 0 {
		if ornil && canBeNil {
			return nil, nil
		} else {
			return nil, ErrValueIsEmpty
		}
	}

	// Check the min length
	if length < min {
		return nil, fmt.Errorf("value is too short, minimum length is %d", min)
	}

	// Check the max length
	if max > min && length > max {
		return nil, fmt.Errorf("value is too long, maximum length is %d", max)
	}

	if res == nil {
		return nil, nil
	}
	return res, nil
}
