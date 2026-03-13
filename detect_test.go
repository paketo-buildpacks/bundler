package bundler_test

import (
	"errors"
	"testing"

	"github.com/paketo-buildpacks/bundler"
	"github.com/paketo-buildpacks/bundler/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		gemfileLockParser *fakes.VersionParser
		detect            packit.DetectFunc
	)

	it.Before(func() {
		gemfileLockParser = &fakes.VersionParser{}

		detect = bundler.Detect(gemfileLockParser)
	})

	it("returns a plan that provides bundler", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: bundler.Bundler},
			},
		}))
	})

	context("when $BP_BUNDLER_VERSION is set and a Gemfile.lock is present", func() {
		it.Before(func() {
			t.Setenv("BP_BUNDLER_VERSION", "1.2.3")
			gemfileLockParser.ParseVersionCall.Returns.Version = "2.1.4"
		})

		it("returns a plan that provides and requires that version of bundler", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: bundler.Bundler},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: bundler.Bundler,
						Metadata: bundler.BuildPlanMetadata{
							VersionSource: "BP_BUNDLER_VERSION",
							Version:       "1.2.3",
						},
					},
					{
						Name: bundler.Bundler,
						Metadata: bundler.BuildPlanMetadata{
							VersionSource: "Gemfile.lock",
							Version:       "2.1.4",
						},
					},
				},
			}))
			Expect(gemfileLockParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/Gemfile.lock"))
		})
	})

	context("when the source code contains a Gemfile.lock file", func() {
		it.Before(func() {
			gemfileLockParser.ParseVersionCall.Returns.Version = "2.1.4"
		})

		it("returns a plan that provides and requires that version of bundler", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: bundler.Bundler},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: bundler.Bundler,
						Metadata: bundler.BuildPlanMetadata{
							VersionSource: "Gemfile.lock",
							Version:       "2.1.4",
						},
					},
				},
			}))
			Expect(gemfileLockParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/Gemfile.lock"))
		})
	})

	context("failure cases", func() {
		context("when the Gemfile.lock parser fails", func() {
			it.Before(func() {
				gemfileLockParser.ParseVersionCall.Returns.Err = errors.New("failed to parse Gemfile.lock")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse Gemfile.lock"))
			})
		})
	})
}
