package bundler

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/bundler-cnb/runner"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBundler(t *testing.T) {
	spec.Run(t, "BundlerRunner", testBundler, spec.Report(report.Terminal{}))
}

func testBundler(t *testing.T, when spec.G, it spec.S) {
	var factory *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)

		factory = test.NewBuildFactory(t)
	})

	when("we are running bundler", func() {
		var fakeRunner *runner.FakeRunner
		var bundle Bundler
		var expectedBundlePath string

		it.Before(func() {
			fakeRunner = &runner.FakeRunner{}
			bundle = NewBundler(factory.Build.Application.Root, "/tmp", factory.Build.Logger)
			bundle.Runner = fakeRunner
			expectedBundlePath = filepath.Join("/tmp", "bundle")
		})

		it("should run bundler -V", func() {
			Expect(bundle.Version()).To(Succeed())
			Expect(fakeRunner.Arguments).To(ConsistOf(expectedBundlePath, "-V"))
		})

		it("should run bundler install", func() {
			Expect(bundle.Install("--foo", "--bar")).To(Succeed())
			Expect(fakeRunner.Arguments).To(ConsistOf(expectedBundlePath, "install", "--quiet", "--foo", "--bar"))
		})
	})

	when("there is a Gemfile in the app root", func() {
		var compsoserPath string
		it.Before(func() {
			compsoserPath = filepath.Join(factory.Build.Application.Root, Gemfile)
			test.WriteFile(t, compsoserPath, "")
		})

		it("should find the Gemfile file", func() {
			path, err := FindGemfile(factory.Build.Application.Root, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(compsoserPath))
		})
	})

	when("there no Gemfile file", func() {
		it("should return an error", func() {
			path, err := FindGemfile(factory.Build.Application.Root, "")
			Expect(path).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no \"" + Gemfile + "\" found"))
		})
	})

	when("there is a Gemfile location specified in buildpack.yml", func() {
		it("should find the Gemfile file under app_root", func() {
			subDir := "subdir"
			test.WriteFile(t, filepath.Join(factory.Build.Application.Root, "buildpack.yml"), `{"bundler": {"gemfile_path": "subdir"}}`)
			compsoserPath := filepath.Join(factory.Build.Application.Root, subDir, Gemfile)
			test.WriteFile(t, compsoserPath, "")
			path, err := FindGemfile(factory.Build.Application.Root, subDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(compsoserPath))
		})
	})

	when("there is a buildpack.yml", func() {
		it("loads and parses with defaults", func() {
			test.WriteFile(t, filepath.Join(factory.Build.Application.Root, "buildpack.yml"), `{"bundler": {"gemfile_path": "subdir"}}`)

			bpYaml, err := LoadBundlerBuildpackYAML(factory.Build.Application.Root)
			Expect(err).ToNot(HaveOccurred())
			Expect(bpYaml.Bundler.GemfilePath).To(Equal("subdir"))
			Expect(bpYaml.Bundler.VendorDirectory).To(Equal("vendor/bundle"))
			Expect(bpYaml.Bundler.InstallOptions).To(ConsistOf([]string{"--without", "development", "test"}))

		})

		it("loads and parses the file", func() {
			test.WriteFile(t, filepath.Join(factory.Build.Application.Root, "buildpack.yml"), `{"bundler": {"gemfile_path": "subdir", "vendor_directory": "somedir", "install_options": ["one", "two", "three"]}}`)

			bpYaml, err := LoadBundlerBuildpackYAML(factory.Build.Application.Root)
			Expect(err).ToNot(HaveOccurred())
			Expect(bpYaml.Bundler.GemfilePath).To(Equal("subdir"))
			Expect(bpYaml.Bundler.VendorDirectory).To(Equal("somedir"))
			Expect(bpYaml.Bundler.InstallOptions).To(ConsistOf("one", "two", "three"))
		})
	})

}
