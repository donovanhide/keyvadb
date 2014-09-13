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
var iterations = flag.Int("iterations", 10000, "number of iterations")
var num = flag.Int("num", 1000, "number of insertions")
var maxDegree = flag.Int("max_degree", 64, "max degree of tree")

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
		degree := uint64(rand.Intn(*maxDegree-2) + 2)
		batch := rand.Intn(int(*num)/10) + 1
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
		p.Title.Text = "Effiency of Batch and Degree"
		p.Add(&EfficiencyScatter{Degree: uint64(*maxDegree), Batch: uint64(*num / 10), Points: points})
		checkErr(p.Save(8, 6, name+".svg"))
	}
}
