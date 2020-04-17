package bundler_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNode(t *testing.T) {
	suite := spec.New("bundler", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Clock", testClock)
	suite("Detect", testDetect)
	suite("GemfileLockParser", testGemfileLockParser)
	suite("GemfileParser", testGemfileParser)
	suite("LogEmitter", testLogEmitter)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite("PlanRefinery", testPlanRefinery)
	suite.Run(t)
}
