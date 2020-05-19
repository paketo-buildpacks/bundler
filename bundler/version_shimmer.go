package bundler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/fs"
)

// Bundler has an "auto-upgrade" feature that means that when simply invoking
// `bundle` from the command-line, you may receive a version that is not what
// was installed by this buildpack:
// https://bundler.io/guides/bundler_2_upgrade.html#version-autoswitch.

// In order to override this behavior, we need to invoke the `bundle`
// executable specifying a version number as is outlined here:
// https://stackoverflow.com/questions/4373128/how-do-i-activate-a-different-version-of-a-particular-gem#answer-4373478

const VersionShimTemplate = "%s _%s_ ${@:-}"

type VersionShimmer struct{}

func NewVersionShimmer() VersionShimmer {
	return VersionShimmer{}
}

func (s VersionShimmer) Shim(dir, version string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return fmt.Errorf("failed to shim bundler executables: %w", err)
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("failed to shim bundler executables: %w", err)
		}

		if info.Mode()&0111 == 0 || info.IsDir() {
			continue
		}

		original := filepath.Join(filepath.Dir(file), fmt.Sprintf("_%s", filepath.Base(file)))
		err = fs.Move(file, original)
		if err != nil {
			return fmt.Errorf("failed to move bundler executables: %w", err)
		}

		content := fmt.Sprintf(VersionShimTemplate, original, version)

		err = ioutil.WriteFile(file, []byte(content), 0755)
		if err != nil {
			return fmt.Errorf("failed to rewrite bundler executables: %w", err)
		}
	}

	return nil
}
