package utils

import (
	"bytes"
	"encoding/json"
	"sort"
)

type State []Param

func NewState(params ...Param) State {
	return State(params)
}

func (s State) Len() int      { return len(s) }
func (s State) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s State) Less(i, j int) bool {
	if s[i].Key == s[j].Key {
		return s[i].Value < s[j].Value
	}
	return s[i].Key < s[j].Key
}

func (s State) Has(key string) bool {
	for _, p := range s {
		if p.Key == key {
			return true
		}
	}
	return false
}

func (s State) Get(key string) string {
	for _, p := range s {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

func (s State) Set(key, value string) State {
	for i, p := range s {
		if p.Key == key {
			s[i].Value = value
			return s
		}
	}
	return append(s, Param{Key: key, Value: value})
}

func (s State) Encode() string {
	sort.Sort(s)
	var buf bytes.Buffer
	_ = buf.WriteByte('{')
	for i, p := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		_ = buf.WriteByte('"')
		_, _ = buf.WriteString(p.Key)
		_, _ = buf.WriteString(`":"`)
		_, _ = buf.WriteString(p.Value)
		_ = buf.WriteByte('"')
	}
	_ = buf.WriteByte('}')
	return buf.String()
}

func DecodeState(s string) State {
	var data map[string]string
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return nil
	}
	var state State
	for k, v := range data {
		state = append(state, Param{Key: k, Value: v})
	}
	return state
}
