package main

import (
	"bundler-cnb/bundler"
	"bundler-cnb/gems"
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	fmt.Println("Implement build")
}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	contributor, willContribute, err := gems.NewContributor(context, bundler.Bundler{})
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := contributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildplan.BuildPlan{})


	return 0, fmt.Errorf("not implemented")
}
