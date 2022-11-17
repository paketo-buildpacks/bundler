package internal_test

import (
	"testing"

	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVersionFinder(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		vf     internal.VersionFinder
	)

	it.Before(func() {
		vf = internal.NewVersionFinder()
	})

	context("FindNewVersions", func() {
		context("when input buildpack config contains no versions", func() {
			it("adds the right number of most recent versions for bundler constraints", func() {
				result, err := vf.FindNewVersions(
					cargo.Config{
						Metadata: cargo.ConfigMetadata{
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								{
									ID:         "bundler",
									Constraint: "1.2.*",
									Patches:    2,
								},
								{
									ID:         "bundler",
									Constraint: "2.3.*",
									Patches:    2,
								},
								{
									ID:         "something-else",
									Constraint: "1.1.*",
									Patches:    1,
								},
							},
						},
					},
					[]internal.Release{
						{Version: "2.4.0"},
						{Version: "2.3.4"},
						{Version: "1.2.7"},
						{Version: "1.2.6"},
						{Version: "2.3.3"},
						{Version: "2.3.2"},
						{Version: "1.1.9"},
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal([]internal.Release{
					{Version: "1.2.7"},
					{Version: "1.2.6"},
					{Version: "2.3.4"},
					{Version: "2.3.3"},
				}))
			})
		})
		context("when the input config contains versions", func() {
			it("it returns only versions greater than the known versions", func() {
				result, err := vf.FindNewVersions(
					cargo.Config{
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{Version: "1.2.3"},
								{Version: "1.2.4"},
							},
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								{
									ID:         "bundler",
									Constraint: "1.2.*",
									Patches:    2,
								},
							},
						},
					},
					[]internal.Release{
						{Version: "2.4.0"},
						{Version: "2.3.4"},
						{Version: "1.2.5"},
						{Version: "1.2.4"},
						{Version: "1.2.3"},
						{Version: "1.1.9"},
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal([]internal.Release{
					{Version: "1.2.5"},
				}))
			})
		})
		context("failure cases", func() {
			context("when the dependency constraint isn't valid semver", func() {
				it("returns an error", func() {
					_, err := vf.FindNewVersions(cargo.Config{
						Metadata: cargo.ConfigMetadata{
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								{ID: "bundler", Constraint: "not-semver"},
							},
						},
					}, []internal.Release{})
					Expect(err).To(MatchError(ContainSubstring("improper constraint")))
				})
			})
			context("when the release version isn't valid semver", func() {
				it("returns an error", func() {
					_, err := vf.FindNewVersions(cargo.Config{
						Metadata: cargo.ConfigMetadata{
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								{ID: "bundler", Constraint: "1.2.*"},
							},
						},
					}, []internal.Release{
						{Version: "not-semver"},
					})
					Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
				})
			})
		})
	})

}
