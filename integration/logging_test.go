package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
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
				Execute(name, filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			sequence := []interface{}{
				MatchRegexp(`MRI Buildpack`),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      Gemfile -> \"~> 2.6.0\"",
				"",
				MatchRegexp(`    Selected MRI version \(using Gemfile\): 2\.6\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.6\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "/home/vcap/.gem/ruby/2\.6\.\d+:/layers/org.cloudfoundry.mri/mri/lib/ruby/gems/2\.6\.\d+"`),
				"",
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      Gemfile.lock -> \"1.17.3\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using Gemfile\.lock\): 1\.17\.3`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 1\.17\.3`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/org.cloudfoundry.bundler/bundler"`),
			}

			Expect(GetBuildLogs(logs.String())).To(ContainSequence(sequence), logs.String())
		})
	})
}
