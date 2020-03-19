package main

import (
	"github.com/simonmittag/jabba/logger"
	"github.com/simonmittag/jabba/server"
	"github.com/simonmittag/jabba/stats"
	"github.com/simonmittag/jabba/time"
)

func main() {
	logger.Init()
	time.Init()
	stats.BootStrap()
	server.BootStrap()
}
