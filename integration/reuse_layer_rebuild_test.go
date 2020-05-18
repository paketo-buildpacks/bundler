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

func testReusingLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}
		name         string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}

	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			build := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(mriBuildpack, bundlerBuildpack, buildPlanBuildpack)

			firstImage, logs, err = build.Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(3))

			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))
			Expect(firstImage.Buildpacks[1].Key).To(Equal("paketo-community/bundler"))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("bundler"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"*\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using <unknown>\): 2\.1\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 2\.1\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/paketo-community_bundler/bundler"`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable(), ContainerLogs(firstContainer.ID))

			// Second pack build
			secondImage, logs, err = build.Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(3))

			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))
			Expect(secondImage.Buildpacks[1].Key).To(Equal("paketo-community/bundler"))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("bundler"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"*\"",
				"",
				MatchRegexp(`    Selected Bundler version \(using <unknown>\): 2\.1\.\d+`),
				"",
				"  Reusing cached layer /layers/paketo-community_bundler/bundler",
			))

			secondContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable(), ContainerLogs(secondContainer.ID))

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_bundler/bundler/bin/bundler"))
			Expect(string(content)).To(MatchRegexp(`Bundler version 2\.1\.\d+`))

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_mri/mri/bin/ruby"))
			Expect(string(content)).To(MatchRegexp(`ruby 2\.6\.\d+`))

			Expect(secondImage.Buildpacks[1].Layers["bundler"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[1].Layers["bundler"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			build := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(mriBuildpack, bundlerBuildpack, buildPlanBuildpack)

			firstImage, logs, err = build.Execute(name, filepath.Join("testdata", "gemfile_lock_version"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(3))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))
			Expect(firstImage.Buildpacks[1].Key).To(Equal("paketo-community/bundler"))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("bundler"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      Gemfile.lock -> \"1.17.3\"",
				"      <unknown>    -> \"*\"",
				"",
				"    Selected Bundler version (using Gemfile.lock): 1.17.3",
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 1\.17\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/paketo-community_bundler/bundler"`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable(), ContainerLogs(firstContainer.ID))

			// Second pack build
			secondImage, logs, err = build.Execute(name, filepath.Join("testdata", "different_version_gemfile_lock"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(3))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))
			Expect(secondImage.Buildpacks[1].Key).To(Equal("paketo-community/bundler"))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("bundler"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Bundler Buildpack %s", buildpackVersion),
				"  Resolving Bundler version",
				"    Candidate version sources (in priority order):",
				"      Gemfile.lock -> \"2.1.4\"",
				"      <unknown>    -> \"*\"",
				"",
				"    Selected Bundler version (using Gemfile.lock): 2.1.4",
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Bundler 2\.1\.\d+`),
				MatchRegexp(`      Completed in \d+\.?\d*`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "\$GEM_PATH:/layers/paketo-community_bundler/bundler"`),
			))

			secondContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable(), ContainerLogs(secondContainer.ID))

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_bundler/bundler/bin/bundler"))
			Expect(string(content)).To(MatchRegexp(`Bundler version 2\.1\.\d+`))

			Expect(string(content)).To(ContainSubstring("/layers/paketo-community_mri/mri/bin/ruby"))
			Expect(string(content)).To(MatchRegexp(`ruby 2\.6\.\d+`))

			Expect(secondImage.Buildpacks[1].Layers["bundler"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[1].Layers["bundler"].Metadata["built_at"]))
		})
	})
}
