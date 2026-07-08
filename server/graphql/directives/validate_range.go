package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

func ValidateRange(ctx context.Context, obj any, next graphql.Resolver, min, max float64, ornil bool) (res any, err error) {
	if res, err = next(ctx); err != nil {
		return nil, err
	}
	switch v := res.(type) {
	case nil:
		if ornil {
			return nil, nil
		}
		return nil, ErrValueIsNil
	case int:
		if v < int(min) || v > int(max) {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *int:
		if *v < int(min) || *v > int(max) {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case int64:
		if v < int64(min) || v > int64(max) {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *int64:
		if *v < int64(min) || *v > int64(max) {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case uint:
		if v < uint(min) || v > uint(max) {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *uint:
		if *v < uint(min) || *v > uint(max) {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case uint64:
		if v < uint64(min) || v > uint64(max) {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *uint64:
		if *v < uint64(min) || *v > uint64(max) {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case float32:
		if v < float32(min) || v > float32(max) {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *float32:
		if *v < float32(min) || *v > float32(max) {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case float64:
		if v < min || v > max {
			if ornil && v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	case *float64:
		if *v < min || *v > max {
			if ornil && *v == 0 {
				return nil, nil
			}
			return nil, ErrValueOutOfRange
		}
	default:
		return nil, ErrValueIsNotNumber
	}
	return res, nil
}
