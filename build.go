package bundler

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan
}

//go:generate faux --interface Shimmer --output fakes/shimmer.go
type Shimmer interface {
	Shim(path, version string) error
}

func Build(
	entries EntryResolver,
	dependencies DependencyManager,
	planRefinery BuildPlanRefinery,
	logger LogEmitter,
	clock chronos.Clock,
	versionShimmer Shimmer,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Bundler version")

		entry := entries.Resolve(context.Plan.Entries)
		version, _ := entry.Metadata["version"].(string)
		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			logger.Subprocess("WARNING: Setting the Bundler version through buildpack.yml will be deprecated soon in Bundler Buildpack v1.0.0.")
			logger.Subprocess("Please specify the version through the $BP_BUNDLER_VERSION environment variable instead. See README.md for more information.")
			logger.Break()
		}

		bundlerLayer, err := context.Layers.Get(Bundler)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundlerLayer.Launch = entry.Metadata["launch"] == true
		bundlerLayer.Build = entry.Metadata["build"] == true
		bundlerLayer.Cache = entry.Metadata["build"] == true

		bom := planRefinery.BillOfMaterial(postal.Dependency{
			ID:      dependency.ID,
			Name:    dependency.Name,
			SHA256:  dependency.SHA256,
			Stacks:  dependency.Stacks,
			URI:     dependency.URI,
			Version: dependency.Version,
		})

		cachedSHA, ok := bundlerLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", bundlerLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{bundlerLayer},
			}, nil
		}

		logger.Process("Executing build process")

		err = bundlerLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Installing Bundler %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			err := dependencies.Install(dependency, context.CNBPath, bundlerLayer.Path)
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
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		bundlerLayer.SharedEnv.Append("GEM_PATH", bundlerLayer.Path, ":")

		logger.Environment(bundlerLayer.SharedEnv)

		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{bundlerLayer},
		}, nil
	}
}
