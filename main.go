package main

import (
	"flag"
)

var (
	addr      = flag.String("addr", "127.0.0.1:5950", "voyager tcp server address")
	verbosity = flag.String("level", "warn", "set log level of clandestine default warn")
)

func main() {
	flag.Parse()
	setUpLogs()
	parseConfig()
}
