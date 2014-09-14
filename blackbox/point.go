package main

import (
	"fmt"
	"image/color"
	"math"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/vg"
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

func (p Point) Color(emin, emax, wcmin, wcmax float64) color.Color {
	l := (float64(p.Efficiency) - emin) / (emax - emin)
	r := (float64(p.WorstCase) - wcmin) / (wcmax - wcmin)
	g := 1 - ((float64(p.WorstCase) - wcmin) / (wcmax - wcmin))
	return &color.RGBA{uint8(r * l * 255), uint8(g * l * 255), 0, 255}
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
	Points []Point
}

func (s *EfficiencyScatter) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)
	emin, emax := s.EfficiencyRange()
	wcmin, wcmax := s.WorstCaseRange()
	xmin, xmax, ymin, ymax := s.DataRange()
	xOffset, yOffset := (trX(xmax)-trX(xmin))/vg.Length(xmax-xmin)/2.2, (trY(ymax)-trY(ymin))/vg.Length(ymax-ymin)/2.2
	for _, pt := range s.Points {
		da.SetColor(pt.Color(emin, emax, wcmin, wcmax))
		x := trX(float64(pt.Degree))
		y := trY(float64(pt.Batch))
		var p vg.Path
		p.Move(x-xOffset, y-yOffset)
		p.Line(x+xOffset, y-yOffset)
		p.Line(x+xOffset, y+yOffset)
		p.Line(x-xOffset, y+yOffset)
		p.Close()
		da.Fill(p)
	}
}

func (s *EfficiencyScatter) EfficiencyRange() (emin, emax float64) {
	emin = math.MaxFloat64
	for _, s := range s.Points {
		if float64(s.Efficiency) > emax {
			emax = float64(s.Efficiency)
		}
		if float64(s.Efficiency) < emin {
			emin = float64(s.Efficiency)
		}
	}
	return
}

func (s *EfficiencyScatter) WorstCaseRange() (wcmin, wcmax float64) {
	wcmin = math.MaxFloat64
	for _, s := range s.Points {
		if float64(s.WorstCase) > wcmax {
			wcmax = float64(s.WorstCase)
		}
		if float64(s.WorstCase) < wcmin {
			wcmin = float64(s.WorstCase)
		}
	}
	return
}

func (s *EfficiencyScatter) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin, ymin = math.MaxFloat64, math.MaxFloat64
	for _, s := range s.Points {
		if float64(s.Batch)+0.5 > ymax {
			ymax = float64(s.Batch) + 0.5
		}
		if float64(s.Batch)-0.5 < ymin {
			ymin = float64(s.Batch) - 0.5
		}
		if float64(s.Degree)+0.5 > xmax {
			xmax = float64(s.Degree) + 0.5
		}
		if float64(s.Degree)-0.5 < xmin {
			xmin = float64(s.Degree) - 0.5
		}
	}
	return
}
