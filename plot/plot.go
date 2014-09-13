package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/donovanhide/keyvadb"
	"github.com/dustin/randbo"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
)

var num = flag.Int("num", 1000, "number of values to insert in one batch")
var rounds = flag.Int("rounds", 100, "number of batches")
var entries = flag.Uint64("entries", 8, "number of entries per tree node")
var seed = flag.Int64("seed", 0, "seed for RNG")

type levelData map[string][]*keyvadb.Summary

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()
	data := make(levelData)
	for _, balancer := range keyvadb.Balancers {
		ms := keyvadb.NewMemoryKeyStore()
		mv := keyvadb.NewMemoryValueStore()
		r := randbo.NewFrom(rand.NewSource(*seed))
		gen := keyvadb.NewRandomValueGenerator(10, 50, r)
		tree, err := keyvadb.NewTree(*entries, ms, mv, balancer.Balancer)
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
			summary, err := keyvadb.NewSummary(tree)
			checkErr(err)
			data[balancer.Name] = append(data[balancer.Name], summary)
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
	for name, summaries := range data {
		points := make(plotter.XYs, len(summaries))
		for round, sum := range summaries {
			points[round].X = float64(round)
			points[round].Y = float64(sum.Total.Entries-sum.Total.Synthetics) / float64(sum.Total.Nodes)
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
	for name, summaries := range data {
		last := summaries[len(summaries)-1]
		points := make(plotter.XYs, len(last.Levels))
		for i, level := range last.Levels {
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
