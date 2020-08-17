package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpacks struct {
		Bundler struct {
			Online  string
			Offline string
		}
		BuildPlan struct {
			Online string
		}
		MRI struct {
			Online  string
			Offline string
		}
	}

	Buildpack struct {
		ID   string
		Name string
	}

	Config struct {
		BuildPlan string `json:"build-plan"`
		MRI       string `json:"mri"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.Bundler.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Bundler.Offline, err = buildpackStore.Get.
		WithVersion("1.2.3").
		WithOfflineDependencies().
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.MRI.Online, err = buildpackStore.Get.
		Execute(settings.Config.MRI)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.MRI.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(settings.Config.MRI)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("buildpack.yml", testBuildpackYML)
	suite("gemfile.lock", testGemfileLock)
	suite("Default", testDefault)
	suite("Logging", testLogging)
	suite("Offline", testOffline)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)

	suite.Run(t)
}
