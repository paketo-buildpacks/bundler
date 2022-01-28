package main

import (
	"os"

	"github.com/paketo-buildpacks/bundler"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

func main() {
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	gemfileLockParser := bundler.NewGemfileLockParser()
	logEmitter := bundler.NewLogEmitter(os.Stdout)
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
			logEmitter,
			chronos.DefaultClock,
			versionShimmer,
		),
	)
}
