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

type Mode uint8

const (
	Bootstrap Mode = 1 << iota
	Server
	Validate
	Usage
	Version
)

func main() {
	mode := Bootstrap

	//trap sigkill and other aborts
	go waitForSignal()

	//for Bootstrap
	defer recovery()

	c := flag.String("c", j8a.DefaultConfigFile, "config file location.")
	o := flag.Bool("o", false, "validate config file, then exit.")
	v := flag.Bool("v", false, "print version.")
	h := flag.Bool("h", false, "print usage instructions.")

	flag.Usage = printUsage
	err := parseFlags()
	if err != nil {
		mode = Usage
	} else {
		mode = Server
		if isFlagPassed("c") {
			j8a.ConfigFile = *c
		}
		if *o {
			mode = Validate
		}
		if *v {
			mode = Version
		}
		if *h {
			mode = Usage
		}
	}

	switch mode {
	case Validate:
		j8a.Validate()
	case Server:
		j8a.Boot.Add(1)
		j8a.BootStrap()
	case Usage:
		printUsage()
	case Version:
		printVersion()
	}
}

func printUsage() {
	printVersion()
	fmt.Printf("Usage: j8a [-c] [-o] | [-v] | [-h]\n")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("j8a[%s] %s\n", j8a.Version, j8a.RandomHuttese())
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

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func parseFlags() error {
	return ParseFlagSet(flag.CommandLine, os.Args[1:])
}

func ParseFlagSet(flagset *flag.FlagSet, args []string) error {
	var positionalArgs []string
	for {
		if err := flagset.Parse(args); err != nil {
			return err
		}

		args = args[len(args)-flagset.NArg():]
		if len(args) == 0 {
			break
		}

		positionalArgs = append(positionalArgs, args[0])
		args = args[1:]
	}
	return flagset.Parse(positionalArgs)
}
