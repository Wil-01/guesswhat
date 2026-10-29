package main

import "testing"

var invalidStrokes = []Stroke{
	{Kind: KindLine, From: Point{-1, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 4},
	{Kind: KindLine, From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "red", Size: 4},
	{Kind: KindLine, From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 500},
	{From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 4},
	{Kind: "laser", From: Point{0, 0}, Color: "#000000", Size: 4},
	{Kind: KindSpray, Color: "#000000", Size: 4},
	{Kind: KindSpray, Points: []Point{{0.5, 0.5}, {2, 0}}, Color: "#000000", Size: 4},
	{Kind: KindSpray, Points: make([]Point, maxSprayDots+1), Color: "#000000", Size: 4},
	{Kind: KindFill, From: Point{0.5, 1.5}, Color: "#000000"},
	{Kind: KindFill, From: Point{0.5, 0.5}, Color: "#00000"},
}

func TestValidStrokes(t *testing.T) {
	valid := []Stroke{
		testStroke,
		{Kind: KindSpray, Points: []Point{{0.1, 0.2}, {0.3, 0.4}}, Color: "#ABCDEF", Size: 10},
		{Kind: KindFill, From: Point{1, 1}, Color: "#ffffff", Begin: true},
	}
	for _, s := range valid {
		if !s.valid() {
			t.Errorf("%+v should be valid", s)
		}
	}
	for _, s := range invalidStrokes {
		if s.valid() {
			t.Errorf("%+v should be invalid", s)
		}
	}
}
