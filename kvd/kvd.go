package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/donovanhide/keyvadb"
)

var port = flag.Int("port", 9000, "port to listen on")
var degree = flag.Uint64("degree", 64, "degree of tree")
var batch = flag.Uint64("batch", 1000, "batch size")
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
			log.Printf("Get: %s", hash)
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
			log.Printf("Add: %s Bytes:%d", hash, len(value))
			if err := db.Add(*hash, value); err != nil {
				writeErr(w, err)
			}
		}
	}
}

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	checkErr(err)
	db, err := keyvadb.NewFileDB(*degree, *batch, *balancer, *name)
	checkErr(err)
	for {
		conn, err := ln.Accept()
		checkErr(err)
		go handleConnection(db, conn)
	}
}
