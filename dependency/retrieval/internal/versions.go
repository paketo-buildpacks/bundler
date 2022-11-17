package internal

import (
	"errors"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/cargo"
)

type VersionFinder struct {
	depID string
}

func NewVersionFinder(depID string) VersionFinder {
	return VersionFinder{depID: depID}
}

func (v VersionFinder) FindNewVersions(bpTOML cargo.Config, releases []Release) ([]Release, error) {
	var newVersions []Release
	for _, constraint := range bpTOML.Metadata.DependencyConstraints {
		if constraint.ID != v.depID {
			continue
		}

		svConstraint, err := semver.NewConstraint(constraint.Constraint)
		if err != nil {
			return nil, err
		}

		latestKnownVersion := getLatestKnownVersion(bpTOML.Metadata.Dependencies, svConstraint)

		var newConstraintVersions []Release
		for _, release := range releases {
			svVersion, err := semver.NewVersion(release.Version)
			if err != nil {
				// Some versions published look like '2.2.0.rc.2', which isn't valid SemVer.
				// It's safe to ignore these.
				if errors.Is(err, semver.ErrInvalidSemVer) {
					continue
				}
				return nil, err
			}

			// Assumes that versions are stored in the release index in reverse-chronological order
			if svConstraint.Check(svVersion) && svVersion.Equal(latestKnownVersion) {
				break
			}

			if svConstraint.Check(svVersion) && svVersion.GreaterThan(latestKnownVersion) {
				newConstraintVersions = append(newConstraintVersions, release)
			}
		}

		sort.Slice(newConstraintVersions, func(i, j int) bool {
			jVersion := semver.MustParse(newConstraintVersions[j].Version)
			iVersion := semver.MustParse(newConstraintVersions[i].Version)
			return iVersion.GreaterThan(jVersion)
		})

		if len(newConstraintVersions) < constraint.Patches {
			newVersions = append(newVersions, newConstraintVersions...)
		} else {
			newVersions = append(newVersions, newConstraintVersions[:constraint.Patches]...)
		}
	}

	// return json of new dependency versions?
	return newVersions, nil
}

func getLatestKnownVersion(deps []cargo.ConfigMetadataDependency, constraint *semver.Constraints) *semver.Version {
	latestVersion := semver.MustParse("0.0.0")
	for _, dependency := range deps {
		svVersion := semver.MustParse(dependency.Version)
		if constraint.Check(svVersion) && svVersion.GreaterThan(latestVersion) {
			latestVersion = svVersion
		}
	}
	return latestVersion
}
