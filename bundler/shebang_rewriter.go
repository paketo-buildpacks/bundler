package bundler

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

type ShebangRewriter struct{}

func NewShebangRewriter() ShebangRewriter {
	return ShebangRewriter{}
}

func (sr ShebangRewriter) Rewrite(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Could not read directory: %w", err)
	}

	for _, file := range files {
		fileContents, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return fmt.Errorf("Could not read file: %w", err)
		}

		shebangRegex := regexp.MustCompile(`^#!/.*ruby.*`)
		fileContents = shebangRegex.ReplaceAll(fileContents, []byte("#!/usr/bin/env ruby"))
		if err := ioutil.WriteFile(filepath.Join(dir, file.Name()), fileContents, 0755); err != nil {
			return fmt.Errorf("Could not write file: %w", err)
		}
	}

	return nil
}
