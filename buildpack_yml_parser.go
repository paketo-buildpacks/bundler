package bundler

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
}

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	var buildpack struct {
		Bundler Config `yaml:"bundler"`
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close file: %v\n", err)
		}
	}()

	err = yaml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return "", err
	}

	return buildpack.Bundler.Version, nil
}
