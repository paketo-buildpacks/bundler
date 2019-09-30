package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/bundler-cnb/gems"
	"github.com/cloudfoundry/bundler-cnb/ruby"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
)

func main() {
	fmt.Println("Implement detect")
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

// TODO: implement the following
// 		- install nodjs, npm, and yarn if needed
func runDetect(context detect.Detect) (int, error) {
	// TODO: include to image of cloudfoundry:cnb
	_, err := exec.LookPath("bundle")
	if err != nil {
		_, err := exec.Command("gem", "install", "bundler").Output()
		if err != nil {
			return context.Fail(), fmt.Errorf("can't install bundler")
		}

	}

	gemfile := filepath.Join(context.Application.Root, "Gemfile")
	if exists, err := helper.FileExists(gemfile); err != nil {
		return context.Fail(), fmt.Errorf("error checking filepath %s", gemfile)
	} else if !exists {
		return context.Fail(), fmt.Errorf("unable to find Gemfile in app root")
	}

	var rubyVersion, bundlerVersion string
	// update how these are calculated
	if rubyVersion, err = ruby.GetRubyVersion(gemfile); err != nil {
		return context.Fail(), fmt.Errorf("unable to resolve ruby version %s", err)
	}

	if bundlerVersion, err = bundler.GetBundlerVersion(gemfile); err != nil {
		return context.Fail(), fmt.Errorf("unable to resolve bundler version %s", err)
	}

	return context.Pass(buildplan.Plan{
		Requires: []buildplan.Required{
			{
				Name:     ruby.Dependency,
				Version:  rubyVersion,
				Metadata: buildplan.Metadata{"build": true, "launch": true},
			},
			{
				Name:     bundler.Dependency,
				Version:  bundlerVersion,
				Metadata: buildplan.Metadata{"build": true, "launch": true},
			},
			{
				Name:     gems.Dependency,
				Metadata: buildplan.Metadata{"launch": true},
			},
		},
		Provides: []buildplan.Provided{
			{Name: bundler.Dependency},
			{Name: gems.Dependency},
		},
	})
}
