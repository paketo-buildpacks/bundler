package gems_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/bundler-cnb/gems"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=gems.go -destination=mocks_test.go -package=gems_test

func TestUnitGems(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Gems", testGems, spec.Report(report.Terminal{}))
}

func testGems(t *testing.T, when spec.G, it spec.S) {
	when("modules.NewContributor", func() {
		var (
			mockCtrl       *gomock.Controller
			mockPkgManager *MockPackageManager
			factory        *test.BuildFactory
		)

		it.Before(func() {
			mockCtrl = gomock.NewController(t)
			mockPkgManager = NewMockPackageManager(mockCtrl)

			factory = test.NewBuildFactory(t)
		})

		it.After(func() {
			mockCtrl.Finish()
		})

		when("there is no Gemfile.lock", func() {
			it("fails", func() {
				factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{})

				_, _, err := gems.NewContributor(factory.Build, mockPkgManager)
				Expect(err).To(HaveOccurred())
			})
		})

		when("there is a Gemfile", func() {
			it.Before(func() {
				test.WriteFile(
					t,
					filepath.Join(factory.Build.Application.Root, "Gemfile"),
					"gemfile contents",
				)
			})

			it("returns true if a build plan exists", func() {
				factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{})

				_, willContribute, err := gems.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(willContribute).To(BeTrue())
			})

			it("returns false if a build plan does not exist", func() {
				_, willContribute, err := gems.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(willContribute).To(BeFalse())
			})

			it("uses Gemfile for identity", func() {
				factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{})

				contributor, _, _ := gems.NewContributor(factory.Build, mockPkgManager)
				name, version := contributor.Metadata.Identity()
				Expect(name).To(Equal(gems.Dependency))
				Expect(version).To(Equal("f1f1324fc1e757d0f2901b6fe7daf8f5cfdc45eaf025cf30b538589951ca78a9"))
			})

			// Gems are operating system independent
			// this should just move the vendored dir to the correct location and set up the env accordingly
			//when("the app is vendored", func() {
			//	it.Before(func() {
			//		test.WriteFile(
			//			t,
			//			filepath.Join(factory.Build.Application.Root, "vendored", "test_module"),
			//			"some module",
			//		)
			//	})
			//
			//	it("contributes gems to the cache layer when included in the build plan", func() {
			//		factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{
			//			Metadata: buildplan.Metadata{"build": true},
			//		})
			//
			//		contributor, _, err := gems.NewContributor(factory.Build, mockPkgManager)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		Expect(contributor.Contribute()).To(Succeed())
			//
			//		layer := factory.Build.Layers.Layer(gems.Dependency)
			//		Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			//		Expect(filepath.Join(layer.Root, "test_module")).To(BeARegularFile())
			//		Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_PATH", layer.Root))
			//
			//		Expect(filepath.Join(factory.Build.Application.Root, "node_gems")).NotTo(BeADirectory())
			//	})
			//
			//	it("contributes gems to the launch layer when included in the build plan", func() {
			//		factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{
			//			Metadata: buildplan.Metadata{"launch": true},
			//		})
			//
			//		contributor, _, err := gems.NewContributor(factory.Build, mockPkgManager)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		Expect(contributor.Contribute()).To(Succeed())
			//
			//		Expect(factory.Build.Layers).To(test.HaveLaunchMetadata(layers.Metadata{Processes: []layers.Process{{"web", "npm start"}}}))
			//
			//		layer := factory.Build.Layers.Layer(gems.Dependency)
			//		Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			//		Expect(filepath.Join(layer.Root, "test_module")).To(BeARegularFile())
			//		Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_PATH", layer.Root))
			//
			//		Expect(filepath.Join(factory.Build.Application.Root, "node_gems")).NotTo(BeADirectory())
			//	})
			//})

			//when("the app is not vendored", func() {
			//
			//	it("contributes gems to the cache layer when included in the build plan", func() {
			//		factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{
			//			Metadata: buildplan.Metadata{"build": true},
			//		})
			//
			//		contributor, _, err := gems.NewContributor(factory.Build, mockPkgManager)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		Expect(contributor.Contribute()).To(Succeed())
			//
			//		layer := factory.Build.Layers.Layer(gems.Dependency)
			//
			//		defaultGemPath := filepath.Join(layer.Root)
			//		mockPkgManager.EXPECT().Install(layer).Do(func(location string) {
			//			test.WriteFile(
			//				t,
			//				filepath.Join(layer, "Gemfile"), // this should be where gems are written eg GEM_PATH?
			//				"some module",
			//			)
			//		})
			//
			//		Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			//		Expect(filepath.Join(layer.Root, "test_module")).To(BeARegularFile()) // TODO: change this to be correct
			//		// Override env variables
			//		environmentDefaults := map[string]string{
			//			"RAILS_ENV":      "production",
			//			"RACK_ENV":       "production",
			//			"RAILS_GROUPS":   "assets",
			//			"BUNDLE_WITHOUT": "development:test",
			//			"BUNDLE_GEMFILE": "Gemfile",
			//			"BUNDLE_BIN":     filepath.Join(layer.Root, "binstubs"),
			//			"BUNDLE_CONFIG":  filepath.Join(layer.Root, "bundle_config"),
			//			"GEM_HOME":       filepath.Join(layer.Root, "gem_home"),
			//			"GEM_PATH": strings.Join([]string{
			//				filepath.Join(layer.Root, "gem_home"),
			//				filepath.Join(layer.Root, "bundler"),
			//			}, ":"),
			//		}
			//
			//		for key, value := range environmentDefaults {
			//			Expect(layer).To(test.HaveOverrideSharedEnvironment(key, value))
			//		}
			//
			//		Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_PATH", layer.Root))
			//
			//		Expect(filepath.Join(factory.Build.Application.Root, "node_gems")).NotTo(BeADirectory())
			//	})
			//
			//	it("contributes gems to the launch layer when included in the build plan", func() {
			//		factory.AddBuildPlan(gems.Dependency, buildplan.Dependency{
			//			Metadata: buildplan.Metadata{"launch": true},
			//		})
			//
			//		contributor, _, err := gems.NewContributor(factory.Build, mockPkgManager)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		Expect(contributor.Contribute()).To(Succeed())
			//
			//		// TODO: add start command
			//		Expect(factory.Build.Layers).To(test.HaveLau(layers.Metadata{Processes: []layers.Process{{"", ""}}}))
			//
			//		layer := factory.Build.Layers.Layer(gems.Dependency)
			//		Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			//		Expect(filepath.Join(layer.Root, "test_module")).To(BeARegularFile())
			//		Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_PATH", layer.Root))
			//
			//		Expect(filepath.Join(factory.Build.Application.Root, "node_gems")).NotTo(BeADirectory())
			//	})
			//})
		})
	})
}
