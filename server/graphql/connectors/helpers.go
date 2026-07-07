package connectors

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

// PtrSlice returns a slice of pointers to each element (for CollectionConnection list storage).
func PtrSlice[T any](list []T) []*T {
	if len(list) == 0 {
		return nil
	}
	out := make([]*T, len(list))
	for i := range list {
		out[i] = &list[i]
	}
	return out
}
