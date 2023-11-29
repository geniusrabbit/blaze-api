package xtypes

// Map type extended with banch of processing methods
type Map[K comparable, V any] map[K]V

// AmpApply the function to each element of the map
func MapApply[K comparable, V any, NK comparable, NV any](mp map[K]V, apply func(key K, val V) (NK, NV)) Map[NK, NV] {
	nMap := make(Map[NK, NV], len(mp))
	for key, val := range mp {
		nkey, nval := apply(key, val)
		nMap[nkey] = nval
	}
	return nMap
}

// MapReduce map and return new value
func MapReduce[K comparable, V any, R any](mp map[K]V, reduce func(key K, val V, ret *R)) R {
	ret := new(R)
	for key, val := range mp {
		reduce(key, val, ret)
	}
	return *ret
}

// MapEqual comparing two maps of the same type
func MapEqual[T ~map[K]V, K comparable, V comparable](mp, otherMap T) bool {
	if len(mp) != len(otherMap) {
		return false
	}
	for key, val := range mp {
		if otherMap[key] != val {
			return false
		}
	}
	return true
}

// Filter map values and return new map without excluded values
func (mp Map[K, V]) Filter(filter func(key K, val V) bool) Map[K, V] {
	nMap := make(Map[K, V], len(mp))
	for key, val := range mp {
		if filter(key, val) {
			nMap[key] = val
		}
	}
	return nMap
}

// Apply the function to each element of the slice
func (mp Map[K, V]) Apply(apply func(key K, val V) (K, V)) Map[K, V] {
	return MapApply(mp, apply)
}

// ReduceIntoOne map and return single value
func (mp Map[K, V]) ReduceIntoOne(reduce func(key K, val V, ret *V)) V {
	return MapReduce(mp, reduce)
}

// Each iterates every element in the map
func (mp Map[K, V]) Each(iter func(key K, val V)) Map[K, V] {
	for key, val := range mp {
		iter(key, val)
	}
	return mp
}

// Copy map
func (mp Map[K, V]) Copy() Map[K, V] {
	nMap := make(Map[K, V], len(mp))
	for key, val := range mp {
		nMap[key] = val
	}
	return nMap
}

// Set value to the map
func (mp Map[K, V]) Set(key K, val V) Map[K, V] {
	mp[key] = val
	return mp
}

// Keys returns all keys of the map
func (mp Map[K, V]) Keys() Slice[K] {
	keys := make([]K, 0, len(mp))
	for key := range mp {
		keys = append(keys, key)
	}
	return keys
}

// Values returns all values of the map
func (mp Map[K, V]) Values() Slice[V] {
	values := make([]V, 0, len(mp))
	for _, val := range mp {
		values = append(values, val)
	}
	return values
}
