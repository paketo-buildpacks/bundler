package internal_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal"
	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal/fakes"
	"github.com/sclevine/spec"
)

func testMetadataGenerator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cs  *fakes.Checksummer
		pg  *fakes.PackageURLGenerator
		gen internal.MetadataGenerator
	)

	context("Generate", func() {
		it.Before(func() {
			cs = &fakes.Checksummer{}
			cs.SumCall.Returns.String = "some-checksum"
			pg = &fakes.PackageURLGenerator{}
			pg.GenerateCall.Returns.String = "some-purl"
			gen = internal.NewMetadataGenerator(cs, pg)
		})

		it("generates a ReleaseMetadata type with the expected values", func() {
			metadata, err := gen.Generate(internal.Release{
				Version:  "1.2.3",
				SHA256:   "abcdef",
				Licenses: []string{"SomeLicense"},
			},
				[]string{"some.stack", "other.stack"},
				"some-target",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(metadata).To(Equal(internal.ReleaseMetadata{
				CPE:             "cpe:2.3:a:bundler:bundler:1.2.3:*:*:*:*:ruby:*:*",
				Licenses:        []string{"SomeLicense"},
				Name:            "bundler",
				ID:              "bundler",
				PURL:            "some-purl",
				SourceChecksum:  "sha256:abcdef",
				SourceURI:       "https://rubygems.org/downloads/bundler-1.2.3.gem",
				Stacks:          []string{"some.stack", "other.stack"},
				StripComponents: 2,
				Target:          "some-target",
				Version:         "1.2.3",
			}))

		})
	})
}
