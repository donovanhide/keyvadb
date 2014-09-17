package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"github.com/donovanhide/keyvadb"
	"github.com/dustin/randbo"
)

var num = flag.Int("num", 10000, "number of values to insert in one batch")
var rounds = flag.Int("rounds", 100, "number of batches")
var degree = flag.Uint64("degree", 9, "number of children per node")
var seed = flag.Int64("seed", 0, "seed for RNG")
var profile = flag.Bool("profile", false, "enable profiling")
var sample = flag.Int("sample", 1, "sampling rate: (round%sample==0)")

type levelData map[string][]*keyvadb.Summary

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()
	if *profile {
		f, err := os.Create("cpu.out")
		checkErr(err)
		pprof.StartCPUProfile(f)
	}
	data := make(levelData)
	for _, balancer := range keyvadb.Balancers {
		ms := keyvadb.NewMemoryKeyStore()
		r := randbo.NewFrom(rand.NewSource(*seed))
		gen := keyvadb.NewRandomValueGenerator(10, 50, r)
		tree, err := keyvadb.NewTree(*degree, ms, balancer.Balancer)
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
			if i%*sample == 0 {
				summary, err := keyvadb.NewSummary(tree)
				checkErr(err)
				data[balancer.Name] = append(data[balancer.Name], summary)
			}
		}
	}
	if *profile {
		pprof.StopCPUProfile()
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
		for i, sum := range summaries {
			round := i * (*sample)
			points[i].X = float64(round)
			points[i].Y = float64(sum.Total.Entries-sum.Total.Synthetics) / float64(sum.Total.Nodes)
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
	description := fmt.Sprintf("%s: %d rounds of %d keys inserted in a tree with %d children per node", title, *rounds, *num, *degree)
	p.Title.Text = description
	filename := strings.Replace(strings.ToLower(title), " ", "_", -1) + ".svg"
	return p.Save(10, 6, filename)
}
