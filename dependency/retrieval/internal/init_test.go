package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitRetrieval(t *testing.T) {
	suite := spec.New("retrieval", spec.Report(report.Terminal{}))
	suite("ReleaseFetcher", testReleaseFetcher)
	suite("MetadataGenerator", testMetadataGenerator)
	suite("VersionFinder", testVersionFinder)
	suite.Run(t)
}
