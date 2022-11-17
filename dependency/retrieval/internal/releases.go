package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Release struct {
	Version    string `json:"number"`
	Licenses   []string
	SHA256     string `json:"sha"`
	Prerelease bool
}

type ReleaseFetcher struct {
	releaseIndex string
}

func NewReleaseFetcher(index string) ReleaseFetcher {
	return ReleaseFetcher{
		releaseIndex: index,
	}
}

func (r ReleaseFetcher) Get() ([]Release, error) {
	response, err := http.Get(r.releaseIndex)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("release index returned status %d", response.StatusCode)
	}

	var releases []Release

	dec := json.NewDecoder(response.Body)

	// read open bracket
	_, err = dec.Token()
	if err != nil {
		return nil, err
	}

	var i int
	for dec.More() {
		var r Release
		err = dec.Decode(&r)
		if err != nil {
			return nil, fmt.Errorf("error parsing release index JSON: %w", err)
		}
		if r.Version == "" {
			return nil, fmt.Errorf("release index element %d missing version", i)
		}
		if r.SHA256 == "" {
			return nil, fmt.Errorf("release index element %d missing sha256", i)
		}
		if r.Prerelease {
			i++
			continue
		}
		releases = append(releases, r)
		i++
	}

	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, errors.New("no valid releases found")
	}

	return releases, nil
}
