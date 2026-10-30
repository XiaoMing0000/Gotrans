package x11win

import "testing"

func TestOpacityToCardinal(t *testing.T) {
	cases := []struct {
		in   float32
		want uint32
	}{
		{1.0, 0xffffffff}, // 100% 必须映射为完全不透明，而不是回绕成 0
		{0.0, 0x00000000},
		{0.5, 0x7fffffff},
		{0.25, 0x3fffffff},
		{1.5, 0xffffffff},  // 超上限钳制
		{-0.5, 0x00000000}, // 超下限钳制
	}
	for _, c := range cases {
		if got := opacityToCardinal(c.in); got != c.want {
			t.Errorf("opacityToCardinal(%v) = %#x, want %#x", c.in, got, c.want)
		}
	}
}
