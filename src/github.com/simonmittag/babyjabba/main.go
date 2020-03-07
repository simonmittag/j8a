package main

import (
	"github.com/simonmittag/babyjabba/logger"
	"github.com/simonmittag/babyjabba/server"
	"github.com/simonmittag/babyjabba/stats"
	"github.com/simonmittag/babyjabba/time"
)

func main() {
	logger.Init()
	time.Init()
	stats.BootStrap()
	server.BootStrap()
}
