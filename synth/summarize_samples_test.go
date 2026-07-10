package synth

import (
	"reflect"
	"testing"
)

func TestSummarizeSamples(t *testing.T) {
	tests := []struct {
		name   string
		tracks [][]float64
		want   []int16
	}{
		{
			// 負のピーク（-2.0）の方が大きい波形。絶対値で正規化されること。
			name:   "negative peak larger than positive",
			tracks: [][]float64{{0.5, -2.0, 1.0}},
			want:   []int16{8191, -32767, 16383},
		},
		{
			name:   "positive peak larger than negative",
			tracks: [][]float64{{2.0, -1.0}},
			want:   []int16{32767, -16383},
		},
		{
			// トラック合算後の値で正規化されること
			name:   "sum across tracks",
			tracks: [][]float64{{1.0, -1.0}, {1.0, -3.0}},
			want:   []int16{16383, -32767},
		},
		{
			// 全サンプル無音でも 0 除算（NaN）にならないこと
			name:   "all silence",
			tracks: [][]float64{{0, 0, 0}},
			want:   []int16{},
		},
		{
			name:   "no tracks",
			tracks: nil,
			want:   []int16{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeSamples(tt.tracks)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("summarizeSamples() = %v, want %v", got, tt.want)
			}
		})
	}
}
