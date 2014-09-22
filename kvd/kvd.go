package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/golang/glog"

	"github.com/donovanhide/keyvadb"
)

var port = flag.Int("port", 9000, "port to listen on")
var degree = flag.Uint64("degree", 84, "degree of tree")
var batch = flag.Uint64("batch", 10000, "batch size")
var cache = flag.Uint64("cache", 4, "number of levels of tree to cache (4 is around 2.4GB)")
var name = flag.String("name", "db", "name of database")
var balancer = flag.String("balancer", "Distance", "balancer to use")

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func writeErr(w *bufio.Writer, err error) {
	log.Println(err)
	_, err = w.WriteString(err.Error() + "\n")
	if err != nil {
		log.Println(err)
	}
	w.Flush()
}

func handleConnection(db *keyvadb.DB, conn net.Conn) {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	for line, err := r.ReadString('\n'); err == nil; line, err = r.ReadString('\n') {
		parts := strings.Split(line[:len(line)-1], ":")
		switch {
		case len(parts) == 1 && parts[0] == "dump":
			count := 0
			err := db.All(func(kv *keyvadb.KeyValue) {
				w.WriteString(kv.String() + "\n")
				count++
			})
			if err != nil {
				w.WriteString(err.Error())
			} else {
				w.WriteString(fmt.Sprintf("End of dump: %d", count))
			}
			w.WriteByte('\n')
			w.Flush()
		case len(parts) == 1 && parts[0] == "range":
			count := 0
			err := db.Range(keyvadb.FirstHash, keyvadb.LastHash, func(kv *keyvadb.KeyValue) {
				w.WriteString(kv.String() + "\n")
				count++
			})
			if err != nil {
				w.WriteString(err.Error())
			} else {
				w.WriteString(fmt.Sprintf("End of range: %d", count))
			}
			w.WriteByte('\n')
			w.Flush()
		case len(parts) == 1:
			hash, err := keyvadb.NewHash(parts[0])
			if err != nil {
				writeErr(w, err)
				continue
			}
			kv, err := db.Get(*hash)
			if err != nil {
				writeErr(w, err)
				continue
			}
			glog.V(2).Infof("Get: %s", hash)
			w.WriteString(fmt.Sprintf("%s:%X\n", kv.Hash, kv.Value))
			w.Flush()
		case len(parts) == 2:
			hash, err := keyvadb.NewHash(parts[0])
			if err != nil {
				writeErr(w, err)
				continue
			}
			value, err := hex.DecodeString(parts[1])
			if err != nil {
				writeErr(w, err)
				continue
			}
			glog.V(2).Infof("Add: %s Bytes:%d", hash, len(value))
			if err := db.Add(*hash, value); err != nil {
				writeErr(w, err)
			}
		}
	}
}

func accept(ln net.Listener, db *keyvadb.DB, done chan bool) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
				log.Fatalln(err)
			}
		}
		go handleConnection(db, conn)
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	db, err := keyvadb.NewFileDB(*degree, *cache, *batch, *balancer, *name)
	checkErr(err)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	checkErr(err)
	done := make(chan bool, 1)
	go accept(ln, db, done)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	done <- true
	checkErr(ln.Close())
	checkErr(db.Close())
}
