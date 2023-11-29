package xtypes

// FirstVal returns the first non-nil value.
func FirstVal[T any](v ...*T) *T {
	for _, val := range v {
		if val != nil {
			return val
		}
	}
	return nil
}
