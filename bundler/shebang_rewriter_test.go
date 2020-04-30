package bundler_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/bundler-cnb/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testShebangRewriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		shebangRewriter bundler.ShebangRewriter
	)

	it.Before(func() {
		shebangRewriter = bundler.NewShebangRewriter()
	})

	context("Rewrite", func() {
		var dir string

		it.Before(func() {
			var err error
			dir, err = ioutil.TempDir("", "bin")
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.WriteFile(filepath.Join(dir, "somescript"), []byte("#!/usr/bin/ruby2.5\n\n\n"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(dir, "anotherscript"), []byte("#!//bin/ruby2.6\n\n\n"), 0755)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		it("removes the ruby version from the shebang", func() {
			Expect(shebangRewriter.Rewrite(dir)).To(Succeed())

			fileContents, err := ioutil.ReadFile(filepath.Join(dir, "somescript"))
			Expect(err).ToNot(HaveOccurred())

			secondFileContents, err := ioutil.ReadFile(filepath.Join(dir, "anotherscript"))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(fileContents)).To(HavePrefix("#!/usr/bin/env ruby"))
			Expect(string(secondFileContents)).To(HavePrefix("#!/usr/bin/env ruby"))
		})

		context("error cases", func() {
			context("when the directory can not be read", func() {
				it("errors", func() {
					Expect(os.RemoveAll(dir)).To(Succeed())

					Expect(shebangRewriter.Rewrite(dir)).To(MatchError(ContainSubstring("Could not read directory")))
				})
			})

			context("when a file could not be read", func() {
				it("errors", func() {
					Expect(os.Chmod(filepath.Join(dir, "somescript"), 0000)).To(Succeed())

					Expect(shebangRewriter.Rewrite(dir)).To(MatchError(ContainSubstring("Could not read file")))
				})
			})

			context("when a file could not be written", func() {
				it("errors", func() {
					Expect(os.Chmod(filepath.Join(dir, "somescript"), 0444)).To(Succeed())

					Expect(shebangRewriter.Rewrite(dir)).To(MatchError(ContainSubstring("Could not write file")))
				})
			})

		})
	})
}
