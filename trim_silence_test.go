package main

import (
	"reflect"
	"testing"
)

func TestTrimSilence(t *testing.T) {
	tests := []struct {
		name  string
		input []int16
		want  []int16
	}{
		{
			name:  "no silence",
			input: []int16{1, 2, 3, 4, 5},
			want:  []int16{1, 2, 3, 4, 5},
		},
		{
			name:  "leading silence",
			input: []int16{0, 0, 0, 1, 2, 3},
			want:  []int16{1, 2, 3},
		},
		{
			name:  "trailing silence",
			input: []int16{1, 2, 3, 0, 0, 0},
			want:  []int16{1, 2, 3},
		},
		{
			name:  "both leading and trailing silence",
			input: []int16{0, 0, 1, 2, 3, 0, 0},
			want:  []int16{1, 2, 3},
		},
		{
			name:  "silence in middle preserved",
			input: []int16{0, 1, 0, 0, 2, 0},
			want:  []int16{1, 0, 0, 2},
		},
		{
			name:  "all zeros",
			input: []int16{0, 0, 0, 0},
			want:  []int16{},
		},
		{
			name:  "empty slice",
			input: []int16{},
			want:  []int16{},
		},
		{
			name:  "single non-zero",
			input: []int16{0, 0, 1, 0, 0},
			want:  []int16{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimSilence(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trimSilence() = %v, want %v", got, tt.want)
			}
		})
	}
}
