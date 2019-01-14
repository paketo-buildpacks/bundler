package gems

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"io/ioutil"
	"path/filepath"
)

const Dependency = "gems"

type PackageManager interface {
	Install(location string) error
}

type Metadata struct {
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return Dependency, m.Hash
}

type Contributor struct {
	Metadata           Metadata
	buildContribution  bool
	launchContribution bool
	pkgManager         PackageManager
	app                application.Application
	layer              layers.Layer
	launch             layers.Layers
}

func NewContributor(context build.Build, pkgManager PackageManager) (Contributor, bool, error) {
	plan, shouldUseBundler := context.BuildPlan[Dependency]
	if !shouldUseBundler {
		return Contributor{}, false, nil
	}

	gemFile := filepath.Join(context.Application.Root, "Gemfile")
	if exists, err := helper.FileExists(gemFile); err != nil {
		return Contributor{}, false, err
	} else if !exists {
		return Contributor{}, false, fmt.Errorf(`unable to find "Gemfile"`)
	}

	buf, err := ioutil.ReadFile(gemFile)
	if err != nil {
		return Contributor{}, false, err
	}

	hash := sha256.Sum256(buf)

	contributor := Contributor{
		app:        context.Application,
		pkgManager: pkgManager,
		layer:      context.Layers.Layer(Dependency),
		launch:     context.Layers,
		Metadata:   Metadata{hex.EncodeToString(hash[:])},
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

// TODO: check if the vendored dir needs to be named vendor
func (c Contributor) Contribute() error {
	return c.layer.Contribute(c.Metadata, func(layer layers.Layer) error {
		vendorDir := filepath.Join(c.app.Root, "vendor")

		vendored, err := helper.FileExists(vendorDir)
		if err != nil {
			return fmt.Errorf("unable to stat gemfile: %s", err.Error())
		}

		if vendored {
			c.layer.Logger.Info("using vendored gems")
			// vendored case
		} else {
			// not vendored
			c.layer.Logger.Info("Installing gems")
		}
		return nil
	})
	// set up the env

	// change installation behavior based on caching

	// write start command
}
