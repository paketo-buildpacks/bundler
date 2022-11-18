package internal

import (
	"net/url"

	"github.com/package-url/packageurl-go"
)

//go:generate faux --interface PackageURLGenerator --output fakes/purl_generator.go
type PackageURLGenerator interface {
	Generate(name, version, sha, uri string) string
}

type PURLGenerator struct{}

func NewPURLGenerator() PURLGenerator {
	return PURLGenerator{}
}

func (PURLGenerator) Generate(name, version, sourceSHA, source string) string {
	purl := packageurl.NewPackageURL(
		packageurl.TypeGeneric,
		"",
		name,
		version,
		packageurl.QualifiersFromMap(map[string]string{
			"checksum":     sourceSHA,
			"download_url": source,
		}),
		"",
	)

	// Unescape the path to remove the added `%2F` and other encodings added to
	// the URL by packageurl-go
	// If the unescaping fails, we should still return the path URL with the
	// encodings, packageurl-go has examples with both the encodings and without,
	// we prefer to avoid the encodings when we can for convenience.
	purlString, err := url.PathUnescape(purl.ToString())
	if err != nil {
		return purl.ToString()
	}

	return purlString
}
