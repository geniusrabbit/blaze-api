package utils

import (
	"encoding/base64"
	"sort"

	"github.com/demdxx/gocast/v2"
	"gopkg.in/mgo.v2/bson"
)

// State represents a collection of parameters.
type State []Param

// NewState creates a new State from the given parameters.
func NewState(params ...Param) State {
	return State(params)
}

// Len returns the length of the State (implements sort.Interface).
func (s State) Len() int { return len(s) }

// Swap swaps two elements in the State (implements sort.Interface).
func (s State) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less compares two elements for sorting (implements sort.Interface).
// Sorts by Key first, then by Value if Keys are equal.
func (s State) Less(i, j int) bool {
	if s[i].Key == s[j].Key {
		return s[i].Value < s[j].Value
	}
	return s[i].Key < s[j].Key
}

// Has checks if a parameter with the given key exists in the State.
func (s State) Has(key string) bool {
	for _, p := range s {
		if p.Key == key {
			return true
		}
	}
	return false
}

// Get retrieves the value for a given key, returning an empty string if not found.
func (s State) Get(key string) string {
	for _, p := range s {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// Extend adds or updates a parameter in the State.
func (s State) Extend(key, value string) State {
	for i, p := range s {
		if p.Key == key {
			s[i].Value = value
			return s
		}
	}
	return append(s, Param{Key: key, Value: value})
}

// Encode serializes the State to a base64-encoded BSON string.
func (s State) Encode() string {
	sort.Sort(s)
	data := make(bson.D, 0, len(s))
	for _, p := range s {
		data = append(data, bson.DocElem{Name: p.Key, Value: p.Value})
	}
	b, _ := bson.Marshal(data)
	return base64.URLEncoding.EncodeToString(b)
}

// DecodeState deserializes a base64-encoded BSON string back into a State.
func DecodeState(s string) State {
	srcData, _ := base64.URLEncoding.DecodeString(s)
	var data bson.M
	_ = bson.Unmarshal(srcData, &data)
	var state State
	for k, v := range data {
		state = append(state, Param{Key: k, Value: gocast.Str(v)})
	}
	return state
}
