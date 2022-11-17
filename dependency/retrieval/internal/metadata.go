package internal

import "fmt"

const cpePattern string = `cpe:2.3:a:bundler:bundler:%s:*:*:*:*:ruby:*:*`
const sourceURI string = `https://rubygems.org/downloads/bundler-%s.gem`
const depID string = "bundler"

type ReleaseMetadata struct {
	CPE string `json:"cpe"`
	// TODO: decide whether to have DeprecationDate in the struct at all
	// DeprecationDate string `json:"deprecation_date, omitempty"`
	Licenses       []string `json:"licenses"`
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	PURL           string   `json:"purl"`
	SourceChecksum string   `json:"source-checksum"`
	SourceURI      string   `json:"source"`
	Stacks         []string `json:"stacks"`
	Target         string   `json:"target"`
	Version        string   `json:"version"`
}

type MetadataGenerator struct {
	name             string
	sourceURIPattern string
	purlGenerator    PackageURLGenerator
	checksummer      Checksummer
}

//go:generate faux --interface Checksummer --output fakes/checksummer.go
type Checksummer interface {
	Sum(paths ...string) (string, error)
}

func NewMetadataGenerator(checksummer Checksummer, purl PackageURLGenerator) MetadataGenerator {
	return MetadataGenerator{
		name:             depID,
		sourceURIPattern: sourceURI,
		checksummer:      checksummer,
		purlGenerator:    purl,
	}
}

func (m MetadataGenerator) Generate(r Release, stackIDs []string, target string) (ReleaseMetadata, error) {
	sourceURI := fmt.Sprintf(m.sourceURIPattern, r.Version)
	return ReleaseMetadata{
		Name:           m.name,
		ID:             m.name,
		Version:        r.Version,
		Stacks:         stackIDs,
		SourceURI:      sourceURI,
		SourceChecksum: r.SHA256,
		// DeprecationDate: "", // This information is not published anywhere TODO: verify this
		CPE:      fmt.Sprintf(cpePattern, r.Version),
		PURL:     m.purlGenerator.Generate(m.name, r.Version, r.SHA256, sourceURI),
		Licenses: r.Licenses,
		Target:   target,
	}, nil
}
