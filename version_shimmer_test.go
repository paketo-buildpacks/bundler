package bundler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVersionShimmer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		versionShimmer bundler.VersionShimmer
		dir            string
	)

	it.Before(func() {
		dir = t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "first"), []byte("first"), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(dir, "second"), []byte("second"), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(dir, "third"), []byte("third"), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(dir, "fourth"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		versionShimmer = bundler.NewVersionShimmer()
	})

	context("Shim", func() {
		it("creates version shims for the executables in the given directory", func() {
			err := versionShimmer.Shim(dir, "some-version")
			Expect(err).NotTo(HaveOccurred())

			files, err := filepath.Glob(filepath.Join(dir, "*"))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(dir, "_first"),
				filepath.Join(dir, "_second"),

				filepath.Join(dir, "first"),
				filepath.Join(dir, "second"),
				filepath.Join(dir, "third"),
				filepath.Join(dir, "fourth"),
			}))

			first, err := os.Open(filepath.Join(dir, "_first"))
			Expect(err).NotTo(HaveOccurred())
			defer first.Close()

			content, err := os.ReadFile(first.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("first"))

			info, err := first.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			firstShim, err := os.Open(filepath.Join(dir, "first"))
			Expect(err).NotTo(HaveOccurred())
			defer firstShim.Close()

			content, err = os.ReadFile(firstShim.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(fmt.Sprintf("#!/usr/bin/env sh\nexec %s _some-version_ ${@:-}", filepath.Join(dir, "_first"))))

			info, err = firstShim.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			second, err := os.Open(filepath.Join(dir, "_second"))
			Expect(err).NotTo(HaveOccurred())
			defer second.Close()

			content, err = os.ReadFile(second.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("second"))

			info, err = second.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			secondShim, err := os.Open(filepath.Join(dir, "second"))
			Expect(err).NotTo(HaveOccurred())
			defer secondShim.Close()

			content, err = os.ReadFile(secondShim.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(fmt.Sprintf("#!/usr/bin/env sh\nexec %s _some-version_ ${@:-}", filepath.Join(dir, "_second"))))

			info, err = secondShim.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			third, err := os.Open(filepath.Join(dir, "third"))
			Expect(err).NotTo(HaveOccurred())
			defer third.Close()

			content, err = os.ReadFile(third.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("third"))

			info, err = third.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0644)))

			fourth, err := os.Open(filepath.Join(dir, "fourth"))
			Expect(err).NotTo(HaveOccurred())
			defer fourth.Close()

			info, err = fourth.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.ModeDir | os.FileMode(0755)))
			Expect(info.IsDir()).To(BeTrue())
		})

		context("failure cases", func() {
			context("when the files in the directory cannot be listed", func() {
				it("returns an error", func() {
					err := versionShimmer.Shim("[]", "some-version")
					Expect(err).To(MatchError("failed to shim bundler executables: syntax error in pattern"))
				})
			})

			context("when the directory cannot be written to", func() {
				it.Before(func() {
					Expect(os.Chmod(filepath.Join(dir, "first"), 0111)).To(Succeed())
				})

				it("returns an error", func() {
					err := versionShimmer.Shim(dir, "some-version")
					Expect(err).To(MatchError(ContainSubstring("failed to move bundler executables:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
