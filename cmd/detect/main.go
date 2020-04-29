package main

import (
	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/packit"
)

func main() {
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	gemfileLockParser := bundler.NewGemfileLockParser()

	packit.Detect(bundler.Detect(buildpackYMLParser, gemfileLockParser))
}
