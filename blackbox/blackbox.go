package main

import (
	"flag"
	"log"
	"math/rand"
	"runtime"

	"code.google.com/p/plotinum/plot"

	"github.com/donovanhide/keyvadb"
	"github.com/dustin/randbo"
)

var seed = flag.Int64("seed", 0, "seed for RNG")
var iterations = flag.Int("iterations", 1000, "number of iterations")
var num = flag.Int("num", 10000, "number of insertions")
var minDegree = flag.Int("min_degree", 2, "minimum degree of tree")
var maxDegree = flag.Int("max_degree", 1024, "maximum degree of tree")
var minBatch = flag.Int("min_batch", 10, "minimum batch size")
var maxBatch = flag.Int("max_batch", 100, "minimum batch size")

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	flag.Parse()
	data := make(map[string][]Point)
	for i := 0; i < *iterations; i++ {
		degree := uint64(rand.Intn(*maxDegree-*minDegree) + *minDegree)
		batch := min(rand.Intn(*maxBatch-*minBatch)+*minBatch, *num)
		for _, balancer := range keyvadb.Balancers {
			ms := keyvadb.NewMemoryKeyStore()
			mv := keyvadb.NewMemoryValueStore()
			r := randbo.NewFrom(rand.NewSource(*seed))
			gen := keyvadb.NewRandomValueGenerator(10, 50, r)
			tree, err := keyvadb.NewTree(degree, ms, mv, balancer.Balancer)
			checkErr(err)
			for left := *num; left > 0; left -= batch {
				take := min(batch, left)
				// log.Printf("%s Left: %d Batch: %d Take: %d Degree: %d", balancer.Name, left, batch, take, degree)
				kv, err := gen.Take(int(take))
				checkErr(err)
				keys := kv.Keys()
				keys.Sort()
				_, err = tree.Add(keys)
				checkErr(err)
			}
			sum, err := keyvadb.NewSummary(tree)
			checkErr(err)
			p := NewPoint(degree, uint64(batch), sum)
			log.Printf("%-8s %s", balancer.Name, p)
			data[balancer.Name] = append(data[balancer.Name], p)
			runtime.GC()
		}
	}
	for name, points := range data {
		p, err := plot.New()
		checkErr(err)
		p.X.Label.Text = "Degree"
		p.Y.Label.Text = "Batch"
		p.Title.Text = "Efficiency of Batch and Degree"
		p.Add(&EfficiencyScatter{Points: points})
		checkErr(p.Save(8, 6, name+".svg"))
	}
}
