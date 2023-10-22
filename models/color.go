package models

import (
	"image/color"
	"math"
)

type XYPoint [2]float64
type Gamut [3]XYPoint

// GamutA LivingColors Iris, Bloom, Aura, LightStrips
var GamutA = Gamut{XYPoint{0.704, 0.296}, XYPoint{0.2151, 0.7106}, XYPoint{0.138, 0.08}}

// GamutB Hue A19 bulbs
var GamutB = Gamut{XYPoint{0.675, 0.322}, XYPoint{0.4091, 0.518}, XYPoint{0.167, 0.04}}

// GamutC Hue BR30, A19 (Gen 3), Hue Go, LightStrips plus
var GamutC = Gamut{XYPoint{0.692, 0.308}, XYPoint{0.17, 0.7}, XYPoint{0.153, 0.048}}

func GetGamutForModel(model string) Gamut {
	switch model {
	case "LST001":
	case "LLC005":
	case "LLC006":
	case "LLC007":
	case "LLC010":
	case "LLC011":
	case "LLC012":
	case "LLC013":
	case "LLC014":
		return GamutA
	case "LCT001":
	case "LCT007":
	case "LCT002":
	case "LCT003":
	case "LLM001":
		return GamutB
	case "LCT010":
	case "LCT011":
	case "LCT012":
	case "LCT014":
	case "LCT015":
	case "LCT016":
	case "LLC020":
	case "LST002":
		return GamutC
	}
	return Gamut{XYPoint{1, 0}, XYPoint{0, 1}, XYPoint{0, 0}}
}

func crossProduct(p1, p2 XYPoint) float64 {
	return p1[0]*p2[1] - p1[1]*p2[0]
}

func distance(p1, p2 XYPoint) float64 {
	dx := p1[0] - p2[0]
	dy := p1[1] - p2[1]
	return math.Sqrt(dx*dx + dy*dy)
}

func (ga Gamut) xyFitsInGamut(p XYPoint) bool {
	v1 := XYPoint{ga[1][0] - ga[0][0], ga[1][1] - ga[0][1]}
	v2 := XYPoint{ga[2][0] - ga[0][0], ga[2][1] - ga[0][1]}

	q := XYPoint{p[0] - ga[0][0], p[1] - ga[0][1]}
	s := crossProduct(q, v2) / crossProduct(v1, v2)
	t := crossProduct(v1, q) / crossProduct(v1, v2)

	return s >= 0.0 && t >= 0.0 && s+t <= 1.0
}

func (ga Gamut) closestPointToLine(A, B, P XYPoint) XYPoint {
	AP := XYPoint{P[0] - A[0], P[1] - A[1]}
	AB := XYPoint{B[0] - A[0], B[1] - A[1]}
	ab2 := AB[0]*AB[0] + AB[1]*AB[1]
	apAb := AP[0]*AB[0] + AP[1]*AB[1]
	t := math.Max(math.Min(apAb/ab2, 0), 1)

	return XYPoint{A[0] + AB[0]*t, A[1] + AB[1]*t}
}

func (ga Gamut) closestPointToPoint(p XYPoint) XYPoint {
	pAB := ga.closestPointToLine(ga[0], ga[1], p)
	pAC := ga.closestPointToLine(ga[2], ga[0], p)
	pBC := ga.closestPointToLine(ga[1], ga[2], p)

	dAB := distance(p, pAB)
	dAC := distance(p, pAC)
	dBC := distance(p, pBC)

	lowest := dAB
	closestPoint := pAB
	if dAC < lowest {
		lowest = dAC
		closestPoint = pAC
	}
	if dBC < lowest {
		lowest = dBC
		closestPoint = pBC
	}
	return XYPoint{closestPoint[0], closestPoint[1]}
}

func (ga Gamut) ColorToXY(c color.Color) XYPoint {
	r, g, b, _ := c.RGBA()
	correctColor := func(channel float64) float64 {
		if channel > 0.04045 {
			return math.Pow((channel+0.055)/(1.0+0.055), 2.4)
		}
		return channel / 12.92
	}
	red := correctColor(float64(r / 255))
	green := correctColor(float64(g / 255))
	blue := correctColor(float64(b / 255))

	X := red*0.664511 + green*0.154324 + blue*0.162028
	Y := red*0.283881 + green*0.668433 + blue*0.047685
	Z := red*0.000088 + green*0.072310 + blue*0.986039
	xy := XYPoint{X / (X + Y + Z), Y / (X + Y + Z)}
	if !ga.xyFitsInGamut(xy) {
		return ga.closestPointToPoint(xy)
	}

	return xy
}

func (ga Gamut) XYYToColor(xy XYPoint, brightness float64) color.Color {
	if !ga.xyFitsInGamut(xy) {
		xy = ga.closestPointToPoint(xy)
	}

	Y := brightness
	X := (Y / xy[1]) * xy[0]
	Z := (Y / xy[1]) * (1 - xy[0] - xy[1])

	r := X*1.656492 - Y*0.354851 - Z*0.255038
	g := -X*0.707196 + Y*1.655397 + Z*0.036152
	b := X*0.051713 - Y*0.121364 + Z*1.011530

	correctColor := func(channel float64) float64 {
		if channel <= 0.0031308 {
			return (1.0+0.055)*math.Pow(channel, 1.0/2.4) - 0.055
		}
		return channel * 12.92
	}
	r = math.Max(correctColor(r), 0)
	g = math.Max(correctColor(g), 0)
	b = math.Max(correctColor(b), 0)

	maxComponent := max(r, g, b)
	if maxComponent > 1 {
		r = r / maxComponent
		g = g / maxComponent
		b = b / maxComponent
	}

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}
