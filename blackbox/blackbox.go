package main

import (
	"flag"
	"log"
	"math/rand"
	"runtime"

	"github.com/donovanhide/keyvadb"
	"github.com/dustin/randbo"
)

var num = flag.Int("num", 10000, "number of insertions")
var seed = flag.Int64("seed", 0, "seed for RNG")
var iterations = flag.Int("iterations", 100, "number of iteration")

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
	for i := 0; i < *iterations; i++ {
		degree := uint64(rand.Intn(256) + 1)
		batch := rand.Intn(*num/10) + 1
		for _, balancer := range keyvadb.Balancers {
			ms := keyvadb.NewMemoryKeyStore()
			mv := keyvadb.NewMemoryValueStore()
			r := randbo.NewFrom(rand.NewSource(*seed))
			gen := keyvadb.NewRandomValueGenerator(10, 50, r)
			tree, err := keyvadb.NewTree(degree, ms, mv, balancer.Balancer)
			checkErr(err)
			for left := *num; left > 0; left -= batch {
				take := min(batch, left)
				// log.Printf("%s Left: %d Batch: %d Take: %d", balancer.Name, left, batch, take)
				kv, err := gen.Take(take)
				checkErr(err)
				keys := kv.Keys()
				keys.Sort()
				_, err = tree.Add(keys)
				checkErr(err)
			}
			summary, err := keyvadb.NewSummary(tree)
			checkErr(err)
			log.Printf("%12s %8d %8d %8d", balancer.Name, degree, batch, summary.Total.Nodes)
			runtime.GC()
		}
	}
}
