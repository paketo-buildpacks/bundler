package integration_test

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

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("logs useful information for the user", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).ToNot(HaveOccurred())

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

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected Bundler version \(using <unknown>\): 2\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 2\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
			))
			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "\$GEM_PATH:/layers/%s/bundler"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))
			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "\$GEM_PATH:/layers/%s/bundler"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))
		})

		context("when the BP_LOG_LEVEL env var is set to DEBUG", func() {
			it("logs useful information for the user", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "default_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.MRI.Online,
						settings.Buildpacks.Bundler.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
					"  Resolving Bundler version",
					"    Candidate version sources (in priority order):",
					"      <unknown> -> \"\"",
				))
				Expect(logs).To(ContainLines(
					MatchRegexp(`    Selected Bundler version \(using <unknown>\): 2\.\d+\.\d+`),
				))
				Expect(logs).To(ContainLines(
					"  Getting the layer associated with Bundler:",
					fmt.Sprintf("    /layers/%s/bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				))
				Expect(logs).To(ContainLines(
					"  Executing build process",
					MatchRegexp(`    Installing Bundler 2\.\d+\.\d+`),
					fmt.Sprintf("    Installation path: /layers/%s/bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					MatchRegexp(`    Source URI\: https\:\/\/deps\.paketo\.io\/bundler\/bundler_2\.\d+\.\d+_linux_noarch_bionic_.*\.tgz`),
					MatchRegexp(`      Completed in \d+\.?\d*`),
				))
				Expect(logs).To(ContainLines(
					"  Configuring build environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "\$GEM_PATH:/layers/%s/bundler"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))
				Expect(logs).To(ContainLines(
					"  Configuring launch environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "\$GEM_PATH:/layers/%s/bundler"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))
			})
		})
	})
}
