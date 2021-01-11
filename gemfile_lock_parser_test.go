package bundler_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/paketo-buildpacks/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGemfileLockParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser bundler.GemfileLockParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "Gemfile.lock")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`GEM
  remote: https://rubygems.org/
  specs:

PLATFORMS
  ruby

DEPENDENCIES

RUBY VERSION
   ruby 2.6.3p62

BUNDLED WITH
	 1.2.3`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = bundler.NewGemfileLockParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("ParseVersion", func() {
		it("parses the bundler major version from a Gemfile.lock file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.*.*"))
		})

		context("when the Gemfile.lock file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(BeEmpty())
			})
		})

		context("failure cases", func() {
			context("when the Gemfile.lock cannot be opened", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse Gemfile.lock:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the bundler version is not valid semver", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`GEM
  remote: https://rubygems.org/
  specs:

PLATFORMS
  ruby

DEPENDENCIES

RUBY VERSION
   ruby 2.6.3p62

BUNDLED WITH
	 not semver`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse Gemfile.lock:")))
					Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
				})
			})
		})
	})
}
