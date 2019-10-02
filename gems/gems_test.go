package gems

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGems(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Gems", testGems, spec.Report(report.Terminal{}))
}

func testGems(t *testing.T, when spec.G, it spec.S) {
	var factory *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
	})

	when("NewContributor", func() {
		it.Before(func() {
			bundlerGemfileString := `ruby '~> 2.6.3'`
			bundlerGemfilePath := filepath.Join(factory.Build.Application.Root, bundler.Gemfile)
			test.WriteFile(t, bundlerGemfilePath, bundlerGemfileString)
		})

		when("there is a lock file", func() {
			it("includes a hash of the lock file in the bundler metadata", func() {
				bundlerLockString := `this is a lock file`
				bundlerLockPath := filepath.Join(factory.Build.Application.Root, bundler.GemfileLock)
				test.WriteFile(t, bundlerLockPath, bundlerLockString)

				contributor, willContribute, err := NewContributor(factory.Build, "/tmp")
				Expect(err).NotTo(HaveOccurred())
				Expect(willContribute).To(BeTrue())
				Expect(contributor.bundlerMetadata.Name).To(Equal("Ruby Bundler"))
				Expect(contributor.bundlerMetadata.Hash).To(Equal("fe2ebd62604e50ad1682fb67979fd368375c2347973c47af8b0394a5359e3e08"))
			})
		})
	})
}
