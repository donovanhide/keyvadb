package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/donovanhide/keyva"
	"github.com/dustin/randbo"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
)

var num = flag.Int("num", 10000, "number of values to insert in one batch")
var rounds = flag.Int("rounds", 100, "number of batches")
var entries = flag.Int("entries", 8, "number of entries per tree node")
var seed = flag.Int64("seed", 0, "seed for RNG")

type levelData map[string][]keyva.LevelSlice

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()
	data := make(levelData)
	for _, balancer := range keyva.Balancers {
		ms := keyva.NewMemoryKeyStore()
		mv := keyva.NewMemoryValueStore()
		r := randbo.NewFrom(rand.NewSource(*seed))
		gen := keyva.NewRandomValueGenerator(10, 50, r)
		tree, err := keyva.NewTree(ms, mv, balancer.Balancer)
		checkErr(err)
		sum := 0
		for i := 0; i < *rounds; i++ {
			kv, err := gen.Take(*num)
			checkErr(err)
			keys := kv.Keys()
			keys.Sort()
			n, err := tree.Add(keys)
			checkErr(err)
			sum += n
			log.Printf("Added %d keys using the %s balancer", sum, balancer.Name)
			levels, err := tree.Levels()
			checkErr(err)
			data[balancer.Name] = append(data[balancer.Name], levels)
		}
	}
	checkErr(save(entriesPlot(data)))
	checkErr(save(distributionPlot(data)))
}

func entriesPlot(data levelData) (*plot.Plot, string) {
	p, err := plot.New()
	checkErr(err)
	p.X.Label.Text = "Round"
	p.Y.Label.Text = "Entries per node"
	var pts []interface{}
	for name, levels := range data {
		points := make(plotter.XYs, len(levels))
		for round, level := range levels {
			total := level.Total()
			points[round].X = float64(round)
			points[round].Y = float64(total.Entries-total.Synthetics) / float64(total.Nodes)
		}
		pts = append(pts, []interface{}{name, points}...)
	}
	checkErr(plotutil.AddLinePoints(p, pts...))
	return p, "Average filled entries per node"
}

func distributionPlot(data levelData) (*plot.Plot, string) {
	p, err := plot.New()
	checkErr(err)
	p.X.Label.Text = "Level"
	p.Y.Label.Text = "Nodes"
	var pts []interface{}
	for name, levels := range data {
		last := levels[len(levels)-1]
		points := make(plotter.XYs, len(last))
		for i, level := range last {
			points[i].X = float64(i)
			points[i].Y = float64(level.Nodes)
		}
		pts = append(pts, []interface{}{name, points}...)
	}
	checkErr(plotutil.AddLinePoints(p, pts...))
	return p, "Nodes per level"
}

func save(p *plot.Plot, title string) error {
	description := fmt.Sprintf("%s: %d rounds of %d keys inserted in a tree with %d entries per node", title, *rounds, *num, *entries)
	p.Title.Text = description
	filename := strings.Replace(strings.ToLower(title), " ", "_", -1) + ".svg"
	return p.Save(10, 6, filename)
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}
