package main

import (
	"flag"
	"fmt"
	"github.com/simonmittag/jabba"
)

func main() {
	cfg := flag.String("c", "./jabba.json", "config file location")
	flag.Usage = func() {
		fmt.Printf(`jabba[%s] "a json friendly reverse TLS proxy for APIs"`, jabba.Version)
		fmt.Print("\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	jabba.ConfigFile = *cfg

	jabba.Boot.Add(1)
	jabba.BootStrap()
}
