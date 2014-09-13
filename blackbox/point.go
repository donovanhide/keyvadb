package main

import (
	"fmt"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"

	"image/color"

	"github.com/donovanhide/keyvadb"
)

type Point struct {
	Degree     uint64
	Batch      uint64
	WorstCase  uint64
	Efficiency float64
}

func (p Point) String() string {
	return fmt.Sprintf("%8d %8d %8d %.4f", p.Degree, p.Batch, p.WorstCase, p.Efficiency)
}

func (p Point) Color() color.Color {
	return &color.RGBA{255, 255, 255, uint8(p.Efficiency * 255)}
	// return &color.RGBA{uint8((1 - p.Efficiency) * 255), uint8(p.Efficiency * 255), 0, 255}
}

func NewPoint(degree, batch uint64, sum *keyvadb.Summary) Point {
	return Point{
		Degree:     degree,
		Batch:      batch,
		WorstCase:  uint64(len(sum.Levels)),
		Efficiency: sum.Efficiency(),
	}
}

type EfficiencyScatter struct {
	Degree, Batch uint64
	Points        []Point
}

func (s *EfficiencyScatter) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)
	circle := plot.CircleGlyph{}
	for _, p := range s.Points {
		fmt.Println(p)
		da.SetColor(p.Color())
		x := trX(float64(p.Degree))
		y := trY(float64(p.Batch))
		circle.DrawGlyph(&da, plotter.DefaultGlyphStyle, plot.Pt(x, y))
	}
}

func (s *EfficiencyScatter) DataRange() (xmin, xmax, ymin, ymax float64) {
	return 0, float64(s.Degree), 0, float64(s.Batch)
}
