package gems

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Dependency = "gems"

type Metadata struct {
	Name string
	Hash string
}

func generateRandomHash() [32]byte {
	randBuf := make([]byte, 512)
	rand.Read(randBuf)
	return sha256.Sum256(randBuf)
}

func (m Metadata) Identity() (name string, version string) {
	return Dependency, m.Hash
}

type Contributor struct {
	app                  application.Application
	bundlerLayer         layers.Layer
	bundlerPackagesLayer layers.Layer
	cacheLayer           layers.Layer
	bundlerMetadata      Metadata
	bundler              bundler.Bundler
	bundlerBuildpackYAML bundler.BuildpackYAML
}

func NewContributor(context build.Build, bundlerPath string) (Contributor, bool, error) {
	buildpackYAML, err := bundler.LoadBundlerBuildpackYAML(context.Application.Root)
	if err != nil {
		return Contributor{}, false, err
	}

	path, err := bundler.FindGemfile(context.Application.Root, buildpackYAML.Bundler.GemfilePath)
	if err != nil {
		return Contributor{}, false, err
	}

	bundlerDir := filepath.Dir(path)
	lockPath := filepath.Join(bundlerDir, bundler.GemfileLock)

	var hash [32]byte
	if exists, err := helper.FileExists(lockPath); err != nil {
		return Contributor{}, false, err
	} else if exists {
		buf, err := ioutil.ReadFile(lockPath)
		if err != nil {
			return Contributor{}, false, err
		}

		hash = sha256.Sum256(buf)
	} else {
		hash = generateRandomHash()
	}

	contributor := Contributor{
		app:                  context.Application,
		bundlerLayer:         context.Layers.Layer(bundler.Dependency),
		bundlerPackagesLayer: context.Layers.Layer(bundler.PackagesDependency),
		cacheLayer:           context.Layers.Layer(bundler.CacheDependency),
		bundlerMetadata:      Metadata{"Ruby Bundler", hex.EncodeToString(hash[:])},
		bundler:              bundler.NewBundler(context.Application.Root, bundlerPath, context.Logger),
		bundlerBuildpackYAML: buildpackYAML,
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	randomHash := generateRandomHash()
	if err := c.cacheLayer.Contribute(Metadata{"Ruby Bundler Cache", hex.EncodeToString(randomHash[:])}, func(layer layers.Layer) error { return nil }, layers.Cache); err != nil {
		return err
	}

	if err := c.setAppVendorDir(); err != nil {
		return err
	}

	packagesFlags := []layers.Flag{layers.Launch}

	if err := c.bundlerPackagesLayer.Contribute(c.bundlerMetadata, c.contributeBundlerPackages, packagesFlags...); err != nil {
		return err
	}

	return helper.WriteSymlink(filepath.Join(c.bundlerPackagesLayer.Root, c.bundlerBuildpackYAML.Bundler.VendorDirectory),
		filepath.Join(c.app.Root, c.bundlerBuildpackYAML.Bundler.VendorDirectory))

}

func (c Contributor) setAppVendorDir() error {
	err := os.Setenv("BUNDLE_PATH", filepath.Join(c.bundlerPackagesLayer.Root, c.bundlerBuildpackYAML.Bundler.VendorDirectory))
	if err != nil {
		return err
	}
	return nil
}

func (c Contributor) contributeBundlerPackages(layer layers.Layer) error {
	if err := os.MkdirAll(layer.Root, os.ModePerm); err != nil {
		return err
	}

	return c.bundler.Install(c.bundlerBuildpackYAML.Bundler.InstallOptions...)
}
