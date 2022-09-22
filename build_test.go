package bundler_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/bundler"
	"github.com/paketo-buildpacks/bundler/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		cnbDir    string

		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		versionShimmer    *fakes.Shimmer
		sbomGenerator     *fakes.SBOMGenerator

		clock  chronos.Clock
		buffer *bytes.Buffer

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "bundler",
			Metadata: map[string]interface{}{
				"version-source": "BP_BUNDLER_VERSION",
				"version":        "2.0.x",
				"launch":         true,
				"build":          true,
			},
		}

		// Legacy SBOM
		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			Name:    "Bundler",
			Version: "2.0.1",
		}
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "bundler",
				Metadata: paketosbom.BOMMetadata{
					Version: "bundler-dependency-version",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "bundler-dependency-sha",
					},
					URI: "bundler-dependency-uri",
				},
			},
		}

		// Syft SBOM
		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		clock = chronos.DefaultClock

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)

		versionShimmer = &fakes.Shimmer{}

		build = bundler.Build(
			entryResolver,
			dependencyManager,
			versionShimmer,
			sbomGenerator,
			logEmitter,
			clock,
		)

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			CNBPath: cnbDir,
			Stack:   "some-stack",
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "bundler",
						Metadata: map[string]interface{}{
							"version-source": "BP_BUNDLER_VERSION",
							"version":        "2.0.x",
							"launch":         true,
							"build":          true,
						},
					},
				},
			},
			Platform: packit.Platform{Path: "platform"},
			Layers:   packit.Layers{Path: layersDir},
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs bundler", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("bundler"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "bundler")))

		Expect(layer.SharedEnv).To(Equal(packit.Environment{
			"GEM_PATH.append": filepath.Join(layersDir, "bundler"),
			"GEM_PATH.delim":  ":",
		}))
		Expect(layer.BuildEnv).To(BeEmpty())
		Expect(layer.LaunchEnv).To(BeEmpty())
		Expect(layer.ProcessLaunchEnv).To(BeEmpty())

		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeFalse())
		Expect(layer.Cache).To(BeFalse())

		Expect(layer.Metadata).To(Equal(map[string]interface{}{
			"dependency-sha": "",
		}))

		Expect(layer.SBOM.Formats()).To(Equal([]packit.SBOMFormat{
			{
				Extension: sbom.Format(sbom.CycloneDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.CycloneDXFormat),
			},
			{
				Extension: sbom.Format(sbom.SPDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SPDXFormat),
			},
		}))

		Expect(filepath.Join(layersDir, "bundler")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "bundler",
				Metadata: map[string]interface{}{
					"version-source": "BP_BUNDLER_VERSION",
					"version":        "2.0.x",
					"launch":         true,
					"build":          true,
				},
			},
		}))
		Expect(entryResolver.MergeLayerTypesCall.Receives.String).To(Equal("bundler"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.BuildpackPlanEntrySlice).To(Equal(
			[]packit.BuildpackPlanEntry{
				{
					Name: "bundler",
					Metadata: map[string]interface{}{
						"version-source": "BP_BUNDLER_VERSION",
						"version":        "2.0.x",
						"launch":         true,
						"build":          true,
					},
				},
			}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("bundler"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("2.0.x"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				Name:    "Bundler",
				Version: "2.0.1",
			},
		}))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			Name:    "Bundler",
			Version: "2.0.1",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "bundler")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(versionShimmer.ShimCall.Receives.Path).To(Equal(filepath.Join(layersDir, "bundler", "bin")))
		Expect(versionShimmer.ShimCall.Receives.Version).To(Equal("2.0.1"))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dependency).To(Equal(postal.Dependency{
			Name:    "Bundler",
			Version: "2.0.1",
		}))
		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "bundler")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Bundler version"))
		Expect(buffer.String()).To(ContainSubstring("Selected Bundler version (using BP_BUNDLER_VERSION): "))
		Expect(buffer.String()).NotTo(ContainSubstring("WARNING: Setting the Bundler version through buildpack.yml will be deprecated soon in Bundler Buildpack v2.0.0."))
		Expect(buffer.String()).NotTo(ContainSubstring("Please specify the version through the $BP_BUNDLER_VERSION environment variable instead. See README.md for more information."))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Configuring build environment"))
		Expect(buffer.String()).To(ContainSubstring("Configuring launch environment"))
	})

	context("when the build plan entry includes the build flag", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "bundler",

				Metadata: map[string]interface{}{
					"version-source": "BP_BUNDLER_VERSION",
					"version":        "2.0.x",
					"build":          true,
				},
			}
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the bundler layer as cached", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("bundler"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeFalse())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "bundler",
						Metadata: paketosbom.BOMMetadata{
							Version: "bundler-dependency-version",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "bundler-dependency-sha",
							},
							URI: "bundler-dependency-uri",
						},
					},
				},
			))
		})
	})

	context("when the build plan entry includes the launch flag", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "bundler",
				Metadata: map[string]interface{}{
					"version-source": "BP_BUNDLER_VERSION",
					"version":        "2.0.x",
					"launch":         true,
				},
			}
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the bundler layer as launch", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("bundler"))

			Expect(layer.Build).To(BeFalse())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeFalse())

			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "bundler",
						Metadata: paketosbom.BOMMetadata{
							Version: "bundler-dependency-version",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "bundler-dependency-sha",
							},
							URI: "bundler-dependency-uri",
						},
					},
				},
			))
		})
	})

	context("when there is a dependency cache match", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, "bundler.toml"), []byte("[metadata]\ndependency-sha = \"some-sha\"\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				Name:   "Bundler",
				SHA256: "some-sha", //nolint:staticcheck
			}
		})

		it("exits build process early", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("bundler"))

			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Bundler version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Bundler version (using BP_BUNDLER_VERSION): "))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
		})
	})

	context("when the build plan entry version source is from buildpack.yml", func() {
		it.Before(func() {
			buildContext.Plan.Entries = append(
				buildContext.Plan.Entries,
				packit.BuildpackPlanEntry{
					Name: "bundler",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "1.17.x",
						"launch":         true,
						"build":          true,
					},
				})

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "bundler",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "1.17.x",
					"launch":         true,
					"build":          true,
				},
			}

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				Name:    "Bundler",
				Version: "1.17.x",
			}

			buildContext.BuildpackInfo.Version = "1.2.3"
		})

		it("returns a result that installs bundler with buildpack.yml", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("bundler"))

			Expect(filepath.Join(layersDir, "bundler")).To(BeADirectory())

			Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
				{
					Name: "bundler",
					Metadata: map[string]interface{}{
						"version-source": "BP_BUNDLER_VERSION",
						"version":        "2.0.x",
						"launch":         true,
						"build":          true,
					},
				},
				{
					Name: "bundler",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "1.17.x",
						"launch":         true,
						"build":          true,
					},
				},
			}))
			Expect(entryResolver.MergeLayerTypesCall.Receives.String).To(Equal("bundler"))
			Expect(entryResolver.MergeLayerTypesCall.Receives.BuildpackPlanEntrySlice).To(Equal(
				[]packit.BuildpackPlanEntry{
					{
						Name: "bundler",
						Metadata: map[string]interface{}{
							"version-source": "BP_BUNDLER_VERSION",
							"version":        "2.0.x",
							"launch":         true,
							"build":          true,
						},
					},
					{
						Name: "bundler",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "1.17.x",
							"launch":         true,
							"build":          true,
						},
					},
				}))

			Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
			Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("bundler"))
			Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("1.17.x"))
			Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
				{
					Name:    "Bundler",
					Version: "1.17.x",
				},
			}))

			Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
				Name:    "Bundler",
				Version: "1.17.x",
			}))
			Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
			Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "bundler")))
			Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

			Expect(versionShimmer.ShimCall.Receives.Path).To(Equal(filepath.Join(layersDir, "bundler", "bin")))
			Expect(versionShimmer.ShimCall.Receives.Version).To(Equal("1.17.x"))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Bundler version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Bundler version (using buildpack.yml): "))
			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the Bundler version through buildpack.yml will be deprecated soon in Bundler Buildpack v2.0.0."))
			Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_BUNDLER_VERSION environment variable instead. See README.md for more information."))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Configuring build environment"))
			Expect(buffer.String()).To(ContainSubstring("Configuring launch environment"))
		})

	})

	context("failure cases", func() {
		context("when a dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when a dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})

		context("when the layers directory cannot be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the Bundler layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "bundler", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "bundler"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "bundler"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})
		})

		context("when the layer directory cannot be removed", func() {
			var layerDir string
			it.Before(func() {
				layerDir = filepath.Join(layersDir, bundler.Bundler)
				Expect(os.MkdirAll(filepath.Join(layerDir, "baller"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(layerDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())
				Expect(os.RemoveAll(layerDir)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the version shimmer cannot create version shims", func() {
			it.Before(func() {
				versionShimmer.ShimCall.Returns.Error = errors.New("failed to create version shims")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to create version shims"))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})
	})
}
