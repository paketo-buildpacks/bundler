package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-community/bundler/bundler"
)

func main() {
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	gemfileLockParser := bundler.NewGemfileLockParser()

	packit.Detect(bundler.Detect(buildpackYMLParser, gemfileLockParser))
}
