package main

import (
	"flag"
	"fmt"
	"github.com/simonmittag/j8a"
)

func main() {
	cfgFile := flag.String("c", j8a.DefaultConfigFile, "config file location")
	flag.Usage = func() {
		fmt.Printf(`j8a[%s] "Achuta! j8a [ dʒʌbbʌ ] is a TLS reverse proxy server for JSON APIs written in golang."`, j8a.Version)
		fmt.Print("\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if isFlagPassed("c") {
		j8a.ConfigFile = *cfgFile
	}

	j8a.Boot.Add(1)
	j8a.BootStrap()
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
