package bundler

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitBundlerContributer(t *testing.T) {
	spec.Run(t, "Bundler", testContributor, spec.Report(report.Terminal{}))
}

func testContributor(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("NewContributor", func() {
		var stubBundlerFixture = filepath.Join("testdata", "stub-bundler.gem")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddPlan(buildpackplan.Plan{Name: Dependency})
			f.AddDependency(Dependency, stubBundlerFixture)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("contributes bundler to the build layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddPlan(buildpackplan.Plan{
				Name:     Dependency,
				Metadata: buildpackplan.Metadata{"build": true},
			})
			f.AddDependency(Dependency, stubBundlerFixture)

			bundlerDep, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(bundlerDep.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, false, false))
			Expect(filepath.Join(layer.Root, BundlerGem)).To(BeARegularFile())
			Expect(filepath.Join(layer.Root, "/bin/bundler")).To(BeARegularFile())
			Expect(filepath.Join(layer.Root, "/bin/bundle")).To(BeARegularFile())
		})
	})
}
