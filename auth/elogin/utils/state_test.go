package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	tests := []struct {
		state     map[string]string
		encdeding string
	}{
		{
			state: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			encdeding: `JwAAAAJrZXkxAAcAAAB2YWx1ZTEAAmtleTIABwAAAHZhbHVlMgAA`,
		},
		{
			state: map[string]string{
				"key2": "value2",
				"key1": "value1",
			},
			encdeding: `JwAAAAJrZXkxAAcAAAB2YWx1ZTEAAmtleTIABwAAAHZhbHVlMgAA`,
		},
		{
			state: map[string]string{
				"2": "2",
				"1": "1",
				"3": "3",
			},
			encdeding: `IAAAAAIxAAIAAAAxAAIyAAIAAAAyAAIzAAIAAAAzAAA=`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("encode%d", i+1), func(t *testing.T) {
			var s State
			for k, v := range tt.state {
				s = s.Extend(k, v)
			}
			assert.Equal(t, tt.encdeding, s.Encode())
		})
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("decode:%d", i+1), func(t *testing.T) {
			s := DecodeState(tt.encdeding)
			assert.Equal(t, len(tt.state), s.Len())
			for k, v := range tt.state {
				assert.Equal(t, v, s.Get(k))
			}
		})
	}
}
