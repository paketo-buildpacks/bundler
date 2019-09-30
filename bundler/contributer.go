package bundler

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bundler-cnb/ruby"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

type Metadata struct {
	Name string
	Hash string
}
type Contributor struct {
	BundlerLayer      layers.DependencyLayer
	RubyLayer         layers.Layer
	buildContribution bool
}

func NewContributor(builder build.Build) (Contributor, bool, error) {
	plan, wantDependency, err := builder.Plans.GetShallowMerged(Dependency)
	if err != nil || !wantDependency {
		return Contributor{}, false, err
	}

	deps, err := builder.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, builder.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		BundlerLayer: builder.Layers.DependencyLayer(dep),
		RubyLayer:    builder.Layers.Layer(ruby.Dependency),
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	return contributor, true, nil
}

func (n Contributor) Contribute() error {
	return n.BundlerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		downloadPath := filepath.Join(layer.Root, BundlerGem)
		err := helper.CopyFile(artifact, downloadPath)
		if err != nil {
			return err
		}

		out, err := exec.Command("gem", "install", "--local", downloadPath, "--install-dir", layer.Root).CombinedOutput()
		if err != nil {
			return fmt.Errorf("can't install dependency bundler:%s %s", err, string(out))
		}

		return nil
	}, n.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	flags := []layers.Flag{}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	return flags
}
