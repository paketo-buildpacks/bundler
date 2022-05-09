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

func testDefault(t *testing.T, context spec.G, it spec.S) {
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

	context("when the buildpack is run with pack build", func() {
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

		it("installs with the defaults", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
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
			Eventually(container).Should(Serve(MatchRegexp(`Bundler version 2\.\d+\.\d+`)).OnPort(8080))

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

		context("validating SBOM", func() {
			var (
				container2 occam.Container
				sbomDir    string
			)

			it.Before(func() {
				var err error
				sbomDir, err = os.MkdirTemp("", "sbom")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
				Expect(os.RemoveAll(sbomDir)).To(Succeed())
			})

			it("writes SBOM files to the layer and label metadata", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "default_app"))
				Expect(err).NotTo(HaveOccurred())

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
					WithSBOMOutputDir(sbomDir).
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
				Eventually(container).Should(Serve(MatchRegexp(`Bundler version 2\.\d+\.\d+`)).OnPort(8080))

				Expect(logs).To(ContainLines(
					fmt.Sprintf("  Generating SBOM for /layers/%s/bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				))
				Expect(logs).To(ContainLines(
					"  Writing SBOM in the following format(s):",
					"    application/vnd.cyclonedx+json",
					"    application/spdx+json",
					"    application/vnd.syft+json",
				))

				// check that legacy SBOM is included via metadata
				container2, err = docker.Container.Run.
					WithCommand("cat /layers/config/metadata.toml").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container2.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(And(
					ContainSubstring("[[bom]]"),
					ContainSubstring(`name = "Bundler`),
					ContainSubstring("[bom.metadata]"),
				))

				// check that all required SBOM files are present
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "bundler", "sbom.cdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "bundler", "sbom.spdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "bundler", "sbom.syft.json")).To(BeARegularFile())

				// check an SBOM file to make sure it has an entry for bundler
				contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "bundler", "sbom.cdx.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(`"name": "Bundler"`))
			})
		})
	})
}
