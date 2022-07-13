package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/j8a"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	go waitForSignal()

	//for Bootstrap
	defer recovery()

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

func recovery() {
	if r := recover(); r != nil {
		pid := os.Getpid()
		log.WithLevel(zerolog.FatalLevel).
			Int("pid", pid).
			Msg("exiting...")
		os.Exit(-1)
	}
}

func waitForSignal() {
	defer recovery()
	sig := interruptChannel()
	for {
		select {
		case <-sig:
			panic("os signal")
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

func interruptChannel() chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	return sigs
}
