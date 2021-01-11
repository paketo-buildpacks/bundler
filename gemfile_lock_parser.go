package bundler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
)

type GemfileLockParser struct{}

func NewGemfileLockParser() GemfileLockParser {
	return GemfileLockParser{}
}

func (p GemfileLockParser) ParseVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("failed to parse Gemfile.lock: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "BUNDLED WITH" {
			if scanner.Scan() {
				version, err := semver.NewVersion(strings.TrimSpace(scanner.Text()))
				if err != nil {
					return "", fmt.Errorf("failed to parse Gemfile.lock: %w", err)
				}

				return fmt.Sprintf("%d.*.*", version.Major()), nil
			}
		}
	}

	return "", nil
}
