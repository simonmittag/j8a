package main

import (
	"github.com/simonmittag/jabba"
)

func main() {
	jabba.InitLogger()
	jabba.InitTime()
	jabba.InitStats()
	jabba.BootStrap()
}
