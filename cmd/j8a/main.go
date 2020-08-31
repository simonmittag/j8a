package main

import (
	"flag"
	"fmt"
	"github.com/simonmittag/j8a"
)

func main() {
	cfg := flag.String("c", "./j8a.yml", "config file location")
	flag.Usage = func() {
		fmt.Printf(`j8a[%s] "TLS reverse proxy server for JSON APIs."`, j8a.Version)
		fmt.Print("\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	j8a.ConfigFile = *cfg

	j8a.Boot.Add(1)
	j8a.BootStrap()
}
