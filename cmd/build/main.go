package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
)

func main() {
	logEmitter := bundler.NewLogEmitter(os.Stdout)
	entryResolver := bundler.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := bundler.NewPlanRefinery()
	clock := bundler.NewClock(time.Now)

	packit.Build(bundler.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock))
}
