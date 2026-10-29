package main

const (
	minStrokeSize = 1
	maxStrokeSize = 50
)

func (p Point) valid() bool {
	return p.X >= 0 && p.X <= 1 && p.Y >= 0 && p.Y <= 1
}

func (s Stroke) valid() bool {
	return s.From.valid() && s.To.valid() &&
		s.Size >= minStrokeSize && s.Size <= maxStrokeSize &&
		isHexColor(s.Color)
}

func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, r := range s[1:] {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}
