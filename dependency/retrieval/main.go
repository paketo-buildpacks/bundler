package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/fs"
)

const depID string = "bundler"

type platformTarget struct {
	stacks []string
	target string
	os     string
	arch   string
}

type buildpackTargetsConfig struct {
	Targets []struct {
		OS   string `toml:"os"`
		Arch string `toml:"arch"`
	} `toml:"targets"`
}

func targetNameFromStackID(stackID string) string {
	const prefix = "io.buildpacks.stacks."
	if strings.HasPrefix(stackID, prefix) {
		return strings.TrimPrefix(stackID, prefix)
	}
	return stackID
}

func getPlatformTargets(bpTOMLPath string, config cargo.Config) ([]platformTarget, error) {
	var targetConfig buildpackTargetsConfig
	if _, err := toml.DecodeFile(bpTOMLPath, &targetConfig); err != nil {
		return nil, err
	}

	if len(targetConfig.Targets) == 0 {
		targetConfig.Targets = append(targetConfig.Targets, struct {
			OS   string `toml:"os"`
			Arch string `toml:"arch"`
		}{OS: "linux", Arch: "amd64"})
	}

	platTargets := []platformTarget{}
	for _, stack := range config.Stacks {
		if stack.ID == "" {
			continue
		}

		targetName := targetNameFromStackID(stack.ID)
		for _, t := range targetConfig.Targets {
			if t.OS == "" || t.Arch == "" {
				continue
			}

			platTargets = append(platTargets, platformTarget{
				stacks: []string{stack.ID},
				target: targetName,
				os:     t.OS,
				arch:   t.Arch,
			})
		}
	}

	return platTargets, nil
}

func main() {
	var bpTOML = flag.String("buildpack-toml-path", "", "Path to buildpack.toml with existing dependencies")
	var output = flag.String("output", "", "the path to a file into which an output metadata JSON will be written")
	var releaseIndex = flag.String("release-index", "https://rubygems.org/api/v1/versions/bundler.json", "the release index to search for new versions")

	flag.Parse()

	if *bpTOML == "" {
		log.Fatal("buildpack-toml-path is required")
	}

	if *output == "" {
		log.Fatal("output is required")
	}

	fetcher := internal.NewReleaseFetcher(*releaseIndex)
	availableVersions, err := fetcher.Get()
	if err != nil {
		log.Fatal(err)
	}

	config, err := cargo.NewBuildpackParser().Parse(*bpTOML)
	if err != nil {
		log.Fatal(err)
	}

	platformTargets, err := getPlatformTargets(*bpTOML, config)
	if err != nil {
		log.Fatal(err)
	}

	finder := internal.NewVersionFinder()
	newVersions, err := finder.FindNewVersions(config, availableVersions)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("New versions: %+v", newVersions)
	var allMetadata []internal.ReleaseMetadata
	generator := internal.NewMetadataGenerator(fs.NewChecksumCalculator(), internal.NewPURLGenerator())
	for _, v := range newVersions {
		for _, pt := range platformTargets {
			metadata, err := generator.Generate(v, pt.stacks, pt.target, pt.os, pt.arch)
			if err != nil {
				log.Fatal(err)
			}
			allMetadata = append(allMetadata, metadata)
		}
	}

	bytes, err := json.Marshal(allMetadata)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("metadata output: %s", string(bytes))

	if err = os.WriteFile(*output, bytes, os.ModePerm); err != nil {
		log.Fatal(fmt.Errorf("cannot write to %s: %w", *output, err))
	}

	log.Printf("Wrote metadata to %s\n", *output)
}
