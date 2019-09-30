package ruby

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const Dependency = "ruby"

func GetRubyVersion(gemFile string) (version string, err error) {
	constraints, err := Version(gemFile)
	if err != nil {
		return "", err
	}

	for index, str := range constraints {
		constraints[index] = fmt.Sprintf("'%s'", str)
	}

	return strings.Join(constraints, ","), nil
}

func Version(gemfile string) ([]string, error) {
	dir := filepath.Dir(gemfile)
	code := `
stdout, $stdout = $stdout, $stderr
begin
  def data()
    return Bundler::Dsl.evaluate("Gemfile", 'Gemfile.lock', {}).ruby_version.engine_versions
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
		return []string{}, err
	}
	output := struct {
		Error string   `json:"error"`
		Data  []string `json:"data"`
	}{}
	if err := json.Unmarshal(body, &output); err != nil {
		return []string{}, err
	}
	if output.Error != "" {
		return []string{}, fmt.Errorf("Running ruby: %s", output.Error)
	}
	return output.Data, nil
}
