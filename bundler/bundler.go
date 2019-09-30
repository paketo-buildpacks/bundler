package bundler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const Dependency = "bundler"
const BundlerGem = "bundler.gem"

type Bundler struct{}

func (b Bundler) Install(location string) error {
	return b.run(location, "install")
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
