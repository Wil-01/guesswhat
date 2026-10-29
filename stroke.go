package main

const (
	KindLine  = "line"
	KindSpray = "spray"
	KindFill  = "fill"

	minStrokeSize = 1
	maxStrokeSize = 50
	maxSprayDots  = 100
)

func (p Point) valid() bool {
	return p.X >= 0 && p.X <= 1 && p.Y >= 0 && p.Y <= 1
}

func (s Stroke) valid() bool {
	if !isHexColor(s.Color) {
		return false
	}
	switch s.Kind {
	case KindLine:
		return s.From.valid() && s.To.valid() && s.Size >= minStrokeSize && s.Size <= maxStrokeSize
	case KindSpray:
		if len(s.Points) == 0 || len(s.Points) > maxSprayDots || s.Size < minStrokeSize || s.Size > maxStrokeSize {
			return false
		}
		for _, p := range s.Points {
			if !p.valid() {
				return false
			}
		}
		return true
	case KindFill:
		return s.From.valid()
	}
	return false
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
