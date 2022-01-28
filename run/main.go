package main

import (
	"os"

	"github.com/paketo-buildpacks/bundler"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	gemfileLockParser := bundler.NewGemfileLockParser()
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	versionShimmer := bundler.NewVersionShimmer()

	packit.Run(
		bundler.Detect(
			buildpackYMLParser,
			gemfileLockParser,
		),
		bundler.Build(
			entryResolver,
			dependencyManager,
			logger,
			chronos.DefaultClock,
			versionShimmer,
		),
	)
}
