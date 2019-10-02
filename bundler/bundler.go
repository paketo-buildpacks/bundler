package bundler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bundler-cnb/runner"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"gopkg.in/yaml.v2"
)

const (
	Dependency         = "bundler"
	CacheDependency    = "ruby-bundler-cache"
	PackagesDependency = "ruby-bundler-packages"
	BundlerGem         = "bundler.gem"
	Gemfile            = "Gemfile"
	GemfileLock        = "Gemfile.lock"
	BundlerCmd         = "bundle"
)

type Bundler struct {
	Logger      logger.Logger
	Runner      runner.Runner
	workingDir  string
	bundlerPath string
}

// NewBundler creates a new Bundler runner
func NewBundler(appRoot, bundlerPath string, logger logger.Logger) Bundler {
	return Bundler{
		Logger: logger,
		Runner: runner.BundlerRunner{
			Logger: logger,
		},
		workingDir:  appRoot,
		bundlerPath: filepath.Join(bundlerPath, BundlerCmd),
	}
}

func (b Bundler) Install(args ...string) error {
	args = append([]string{"install", "--quiet"}, args...)
	return b.Runner.Run(b.bundlerPath, b.workingDir, args...)
}

func (b Bundler) run(dir string, args ...string) error {
	cmd := exec.Command("bundle", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetBundlerVersion(gemFile string) (version string, err error) {
	return Version(gemFile)
}

func Version(gemfile string) (string, error) {
	dir := filepath.Dir(gemfile)
	code := `
stdout, $stdout = $stdout, $stderr
begin
  def data()
    v = Bundler::Dsl.evaluate("Gemfile", 'Gemfile.lock', {}).locked_gems.bundler_version.version
    v == "" ? Bundler::VERSION : v
  end
  out = data()
  stdout.puts({error:nil, data:out}.to_json)
rescue => e
  stdout.puts({error:e.to_s, data:nil}.to_json)
end
`

	cmd := exec.Command("ruby", "-rjson", "-rbundler", "-e", code)
	cmd.Dir = dir
	body, err := cmd.Output()
	if err != nil {
		fmt.Println(body)
		return "", err
	}
	output := struct {
		Error string `json:"error"`
		Data  string `json:"data"`
	}{}
	if err := json.Unmarshal(body, &output); err != nil {
		return "", err
	}
	if output.Error != "" {
		return "", fmt.Errorf("Running ruby: %s", output.Error)
	}
	return output.Data, nil
}

// Version runs `bundler version`
func (c Bundler) Version() error {
	return c.Runner.Run(c.bundlerPath, c.workingDir, "-V")
}

// Config runs `bundle config`
func (c Bundler) Config(key, value string) error {
	args := []string{"config", key, value}
	return c.Runner.Run(c.bundlerPath, c.workingDir, args...)
}

type BundlerConfig struct {
	Version         string   `yaml:"version"`
	InstallOptions  []string `yaml:"install_options"`
	VendorDirectory string   `yaml:"vendor_directory"`
	GemfilePath     string   `yaml:"gemfile_path"`
}

type BuildpackYAML struct {
	Bundler BundlerConfig `yaml:"bundler"`
}

// LoadBundlerBuildpackYAML loads the buildpack YAML from disk
func LoadBundlerBuildpackYAML(appRoot string) (BuildpackYAML, error) {
	buildpackYAML, configFile := BuildpackYAML{}, filepath.Join(appRoot, "buildpack.yml")

	buildpackYAML.Bundler.InstallOptions = []string{"--without", "development", "test"}
	buildpackYAML.Bundler.VendorDirectory = "vendor/bundle"

	if exists, err := helper.FileExists(configFile); err != nil {
		return BuildpackYAML{}, err
	} else if exists {
		file, err := os.Open(configFile)
		if err != nil {
			return BuildpackYAML{}, err
		}
		defer file.Close()

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return BuildpackYAML{}, err
		}

		err = yaml.Unmarshal(contents, &buildpackYAML)
		if err != nil {
			return BuildpackYAML{}, err
		}
	}
	return buildpackYAML, nil
}

// FindGemfile locates the Gemfile and Gemfile.lock files
func FindGemfile(appRoot string, gemFilePath string) (string, error) {

	paths := []string{
		filepath.Join(appRoot, Gemfile),
	}

	if gemFilePath != "" {
		paths = append(
			paths,
			filepath.Join(appRoot, gemFilePath, Gemfile),
		)
	}

	for _, path := range paths {
		if exists, err := helper.FileExists(path); err != nil {
			return "", fmt.Errorf("error checking filepath: %s", path)
		} else if exists {
			return path, nil
		}
	}

	return "", fmt.Errorf(`no "%s" found in the following locations: %v`, Gemfile, paths)
}
