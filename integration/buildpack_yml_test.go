package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

			name   string
			source string
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
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("installs the version specified therein", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "buildpack_yml_version"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.Bundler.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				WithCommand("ruby run.rb").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring(fmt.Sprintf("/layers/%s/bundler/bin/bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")))).OnPort(8080))
			Eventually(container).Should(Serve(MatchRegexp(`Bundler version 1\.17\.\d+`)).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.17.x\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using buildpack\.yml\): 1\.17\.\d+`),
				"",
				"    WARNING: Setting the Bundler version through buildpack.yml will be deprecated soon in Bundler Buildpack v1.0.0.",
				"    Please specify the version through the $BP_BUNDLER_VERSION environment variable instead. See README.md for more information.",
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 1\.17\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "\$GEM_PATH:/layers/%s/bundler"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))
		})
	})
}
