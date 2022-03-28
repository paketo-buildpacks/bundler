package bundler

import (
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface Shimmer --output fakes/shimmer.go
type Shimmer interface {
	Shim(path, version string) error
}

func Build(
	entries EntryResolver,
	dependencies DependencyManager,
	logger scribe.Emitter,
	clock chronos.Clock,
	versionShimmer Shimmer,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Bundler version")

		entry, allEntries := entries.Resolve("bundler", context.Plan.Entries, []interface{}{"BP_BUNDLER_VERSION", "buildpack.yml", "Gemfile.lock"})
		logger.Candidates(allEntries)

		version, _ := entry.Metadata["version"].(string)
		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the Bundler version through buildpack.yml will be deprecated soon in Bundler Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_BUNDLER_VERSION environment variable instead. See README.md for more information.")
			logger.Break()
		}

		logger.Debug.Process("Generating the SBOM")
		logger.Debug.Break()
		bom := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entries.MergeLayerTypes("bundler", context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		logger.Debug.Process("Getting the layer associated with Bundler:")
		bundlerLayer, err := context.Layers.Get(Bundler)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(bundlerLayer.Path)
		logger.Debug.Break()

		cachedSHA, ok := bundlerLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
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

		bundlerLayer.Metadata = map[string]interface{}{
			DepKey: dependency.SHA256,
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
