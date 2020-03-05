package main

import (
	"github.com/simonmittag/babyjabba/logger"
	"github.com/simonmittag/babyjabba/server"
	"github.com/simonmittag/babyjabba/time"
)

func main() {
	logger.Init()
	time.Init()
	server.BootStrap()
}
