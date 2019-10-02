package main

import (
	"fmt"

	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/bundler-cnb/gems"
	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	fmt.Println("Implement build")
}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))
	bundlerContributor, willContributeBundler, err := bundler.NewContributor(context)
	if err != nil {
		return context.Failure(102), err
	}

	if willContributeBundler {
		err := bundlerContributor.Contribute()
		if err != nil {
			return context.Failure(103), err
		}

		gemsContributor, willContributeGems, err := gems.NewContributor(context, bundlerContributor.BundlerLayer.Root)
		if err != nil {
			return context.Failure(102), err
		}
		if willContributeGems {
			if err := gemsContributor.Contribute(); err != nil {
				return context.Failure(103), err
			}
		}
	}

	return context.Success()
}
