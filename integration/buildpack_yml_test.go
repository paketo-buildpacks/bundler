package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackYML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the source code contains a buildpack.yml", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("installs the version specified therein", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(mriBuildpack, bundlerBuildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "buildpack_yml_version"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable(), ContainerLogs(container.ID))

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_bundler/bundler/bin/bundler"))
			Expect(string(content)).To(MatchRegexp(`Bundler version 1\.17\.\d+`))

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_mri/mri/bin/ruby"))
			Expect(string(content)).To(MatchRegexp(`ruby 2\.7\.\d+`))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.17.x\"",
				"      <unknown>     -> \"*\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using buildpack\.yml\): 1\.17\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 1\.17\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/paketo-community_bundler/bundler"`),
			))
		})
	})
}
