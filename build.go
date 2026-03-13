package bundler

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface Shimmer --output fakes/shimmer.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

type Shimmer interface {
	Shim(path, version string) error
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func Build(
	dependencies DependencyManager,
	versionShimmer Shimmer,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Bundler version")

		planner := draft.NewPlanner()

		entry, allEntries := planner.Resolve("bundler", context.Plan.Entries, []interface{}{"BP_BUNDLER_VERSION", "Gemfile.lock"})
		logger.Candidates(allEntries)

		version, _ := entry.Metadata["version"].(string)
		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes("bundler", context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = legacySBOM
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = legacySBOM
		}

		logger.Debug.Process("Getting the layer associated with Bundler:")
		bundlerLayer, err := context.Layers.Get(Bundler)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(bundlerLayer.Path)
		logger.Debug.Break()

		cachedChecksum, ok := bundlerLayer.Metadata[DepKey].(string)

		if ok && cargo.Checksum(dependency.Checksum).MatchString(cachedChecksum) {
			logger.Process("Reusing cached layer %s", bundlerLayer.Path)
			logger.Break()

			bundlerLayer.Launch, bundlerLayer.Build, bundlerLayer.Cache = launch, build, build
			return packit.BuildResult{
				Layers: []packit.Layer{bundlerLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		bundlerLayer, err = bundlerLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundlerLayer.Launch, bundlerLayer.Build, bundlerLayer.Cache = launch, build, build

		logger.Subprocess("Installing Bundler %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			logger.Debug.Subprocess("Installation path: %s", bundlerLayer.Path)
			logger.Debug.Subprocess("Source URI: %s", dependency.URI)
			err := dependencies.Deliver(dependency, context.CNBPath, bundlerLayer.Path, context.Platform.Path)
			if err != nil {
				return err
			}

			return versionShimmer.Shim(filepath.Join(bundlerLayer.Path, "bin"), dependency.Version)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(bundlerLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, bundlerLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		bundlerLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundlerLayer.Metadata = map[string]interface{}{
			DepKey: dependency.Checksum,
		}

		bundlerLayer.SharedEnv.Append("GEM_PATH", bundlerLayer.Path, ":")
		logger.EnvironmentVariables(bundlerLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{bundlerLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
