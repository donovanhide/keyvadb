package keyva

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

var debug = flag.Bool("debug", false, "Debug output")

func init() {
	flag.Parse()
}

func dumpWithTitle(title string, data interface{}, indent int) string {
	if reflect.TypeOf(data).Kind() != reflect.Slice {
		panic(fmt.Sprintf("cannot dump %+v", data))
	}
	tab := strings.Repeat("\t", indent)
	s := []string{tab + title}
	slice := reflect.ValueOf(data)
	for i := 0; i < slice.Len(); i++ {
		s = append(s, fmt.Sprintf("%s%06d:%s", tab, i, slice.Index(i).Interface()))
	}
	s = append(s, tab+strings.Repeat("-", len(title)))
	return strings.Join(s, "\n")
}

func debugPrintln(v interface{}) {
	if *debug {
		fmt.Println(v)
	}
}

func debugPrintf(format string, a ...interface{}) {
	if *debug {
		fmt.Printf(format, a...)
	}
}
