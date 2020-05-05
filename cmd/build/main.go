package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
	"github.com/paketo-community/bundler/bundler"
)

func main() {
	logEmitter := bundler.NewLogEmitter(os.Stdout)
	entryResolver := bundler.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := bundler.NewPlanRefinery()
	clock := bundler.NewClock(time.Now)
	shebangRewriter := bundler.NewShebangRewriter()

	packit.Build(bundler.Build(
		entryResolver,
		dependencyManager,
		planRefinery,
		logEmitter,
		clock,
		shebangRewriter,
	))
}
