package main

import (
	"bundler-cnb/bundler"
	"bundler-cnb/gems"
	"bundler-cnb/ruby"
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

// TODO: test case when Gemfile is missing
// TODO: handle case for custom named Gemfile, ex: firstGemFile
func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		factory = test.NewDetectFactory(t)
	})

	when("No Gemfile present", func() {
		it("detection fails", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("Gemfile is present", func() {
		when("Gemfile.lock was bundled with version 1.X.X", func(){
			it.Before(func(){
				GemfileString := fmt.Sprintf(`ruby '~> 3.2', '< 3.2.5'

gem 'uglifier', '>= 1.3.0'`)
				test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gemfile"), GemfileString)

				GemfileLockString := fmt.Sprintf(`GEM
  specs:
    execjs (2.7.0)
    uglifier (3.1.7)
      execjs (>= 0.3.0, < 3)

PLATFORMS
  ruby

DEPENDENCIES
  uglifier (>= 1.3.0)

RUBY VERSION
   ruby 2.4.2p0

BUNDLED WITH
   1.16.4
`)
				test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gemfile.lock"), GemfileLockString)
			})

			it("detection succeeds with bundler and ruby versions from the Gemfile.lock", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					ruby.Dependency: buildplan.Dependency{
						Version:  "'~> 3.2','< 3.2.5'",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					bundler.Dependency: buildplan.Dependency{
						Version:  "1.16.4",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					gems.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					},
				}))
			})
		})
		when("Gemfile.lock was bundled with version 1.X.X", func(){
		it.Before(func(){
			GemfileString := fmt.Sprintf(`ruby '~> 3.2'

gem 'uglifier', '>= 1.3.0'`)
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gemfile"), GemfileString)

			GemfileLockString := fmt.Sprintf(`GEM
  specs:
    execjs (2.7.0)
    uglifier (3.1.7)
      execjs (>= 0.3.0, < 3)

PLATFORMS
  ruby

DEPENDENCIES
  uglifier (>= 1.3.0)

RUBY VERSION
   ruby 2.4.2p0

BUNDLED WITH
   2.0.1
`)
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gemfile.lock"), GemfileLockString)
		})
		it("detection succeeds with bundler and ruby versions from the Gemfile.lock", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Output).To(Equal(buildplan.BuildPlan{
				ruby.Dependency: buildplan.Dependency{
					Version:  "'~> 3.2'",
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				},
				bundler.Dependency: buildplan.Dependency{
					Version:  "2.0.1",
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				},
				gems.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true},
				},
			}))
		})
	})

	})

}
