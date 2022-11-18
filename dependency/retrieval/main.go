package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/fs"
)

const depID string = "bundler"

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

	finder := internal.NewVersionFinder()
	newVersions, err := finder.FindNewVersions(config, availableVersions)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("New versions: %+v", newVersions)
	var allMetadata []internal.ReleaseMetadata
	generator := internal.NewMetadataGenerator(fs.NewChecksumCalculator(), internal.NewPURLGenerator())
	for _, v := range newVersions {
		metadata, err := generator.Generate(v,
			[]string{
				"io.buildpacks.stacks.bionic",
				"io.buildpacks.stacks.jammy",
			},
			"ubuntu")
		if err != nil {
			log.Fatal(err)
		}
		allMetadata = append(allMetadata, metadata)
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
