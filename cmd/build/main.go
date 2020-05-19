package main

import (
	"os"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-community/bundler/bundler"
)

func main() {
	logEmitter := bundler.NewLogEmitter(os.Stdout)
	entryResolver := bundler.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := bundler.NewPlanRefinery()
	clock := bundler.NewClock(time.Now)
	versionShimmer := bundler.NewVersionShimmer()

	packit.Build(bundler.Build(
		entryResolver,
		dependencyManager,
		planRefinery,
		logEmitter,
		clock,
		versionShimmer,
	))
}
