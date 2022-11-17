package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Release struct {
	Version  string `json:"number"` // TODO: handle prereleases
	Licenses []string
	SHA256   string `json:"sha"`
}

type ReleaseFetcher struct {
	releaseIndex string
}

func NewReleaseFetcher() ReleaseFetcher {
	return ReleaseFetcher{
		releaseIndex: "https://rubygems.org/api/v1/versions/bundler.json",
	}
}

func (r ReleaseFetcher) Get() ([]Release, error) {
	response, err := http.Get(r.releaseIndex)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("non 200 status code") // TODO: think more about handling
	}

	var releases []Release

	err = json.NewDecoder(response.Body).Decode(&releases)
	if err != nil {
		return nil, fmt.Errorf("error parsing release JSON") // TODO: don't pull in entire json?
	}

	return releases, nil
}
