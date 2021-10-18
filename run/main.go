package main

import (
	"os"

	"github.com/paketo-buildpacks/bundler"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
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
