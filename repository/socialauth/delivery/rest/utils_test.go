package rest

import "testing"

func TestUrlSetQueryParams(t *testing.T) {
	tests := []struct {
		name   string
		sUrl   string
		params map[string]string
		want   string
	}{
		{
			name:   "empty",
			sUrl:   "",
			params: map[string]string{},
			want:   "",
		},
		{
			name:   "no params",
			sUrl:   "http://example.com",
			params: map[string]string{},
			want:   "http://example.com",
		},
		{
			name:   "no query",
			sUrl:   "http://example.com",
			params: map[string]string{"a": "1", "b": "2"},
			want:   "http://example.com?a=1&b=2",
		},
		{
			name:   "with query",
			sUrl:   "http://example.com?c=3",
			params: map[string]string{"a": "1", "b": "2"},
			want:   "http://example.com?a=1&b=2&c=3",
		},
		{
			name:   "with pattern",
			sUrl:   "http://example.com/{a}",
			params: map[string]string{"a": "1", "b": "2"},
			want:   "http://example.com/1?b=2",
		},
		{
			name:   "with pattern and query",
			sUrl:   "http://example.com/{a}?c=3",
			params: map[string]string{"a": "1", "b": "2"},
			want:   "http://example.com/1?b=2&c=3",
		},
		{
			name:   "with pattern and query and pattern",
			sUrl:   "http://example.com/{a}?c=3&d={b}",
			params: map[string]string{"a": "1", "b": "2", "e": "4"},
			want:   "http://example.com/1?c=3&d=2&e=4",
		},
		{
			name:   "with pattern and query and pattern2",
			sUrl:   "http://example.com/{a}?c=3&d={b}",
			params: map[string]string{"a": "1", "b": "2", "e": "4", "c": "5"},
			want:   "http://example.com/1?c=5&d=2&e=4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlSetQueryParams(tt.sUrl, tt.params); got != tt.want {
				t.Errorf("urlSetQueryParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
