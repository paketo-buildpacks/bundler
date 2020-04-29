package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/occam/matchers"
	. "github.com/onsi/gomega"
)

func testBuildpackYML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image occam.Image
			name  string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("logs useful information for the user", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(mriBuildpack, bundlerBuildpack).
				Execute(name, filepath.Join("testdata", "buildpack_yml_version"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.17.x\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using buildpack\.yml\): 1\.17\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 1\.17\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/org.cloudfoundry.bundler/bundler"`),
			))
		})
	})
}
