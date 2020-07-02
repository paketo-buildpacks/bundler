package integration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	mriBuildpack            string
	offlineMRIBuildpack     string
	bundlerBuildpack        string
	offlineBundlerBuildpack string
	buildPlanBuildpack      string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())

	bundlerBuildpack, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	offlineBundlerBuildpack, _, err = dagger.PackageCachedBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	buildPlanBuildpack, err = dagger.GetLatestCommunityBuildpack("ForestEckhardt", "build-plan")
	Expect(err).ToNot(HaveOccurred())

	mriBuildpack, err = dagger.GetLatestCommunityBuildpack("paketo-community", "mri")
	Expect(err).ToNot(HaveOccurred())

	mriSource, err := dagger.GetLatestUnpackagedCommunityBuildpack("paketo-community", "mri")
	Expect(err).ToNot(HaveOccurred())

	offlineMRIBuildpack, _, err = dagger.PackageCachedBuildpack(mriSource)
	Expect(err).ToNot(HaveOccurred())

	// HACK: we need to fix dagger and the package.sh scripts so that this isn't required
	bundlerBuildpack = fmt.Sprintf("%s.tgz", bundlerBuildpack)
	offlineBundlerBuildpack = fmt.Sprintf("%s.tgz", offlineBundlerBuildpack)
	offlineMRIBuildpack = fmt.Sprintf("%s.tgz", offlineMRIBuildpack)

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("buildpack.yml", testBuildpackYML)
	suite("gemfile.lock", testGemfileLock)
	suite("Default", testDefault)
	suite("Logging", testLogging)
	suite("Offline", testOffline)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)

	defer AfterSuite(t)
	suite.Run(t)
}

func AfterSuite(t *testing.T) {
	var Expect = NewWithT(t).Expect

	Expect(dagger.DeleteBuildpack(mriBuildpack)).To(Succeed())
	Expect(dagger.DeleteBuildpack(offlineMRIBuildpack)).To(Succeed())
	Expect(dagger.DeleteBuildpack(bundlerBuildpack)).To(Succeed())
	Expect(dagger.DeleteBuildpack(offlineBundlerBuildpack)).To(Succeed())
	Expect(dagger.DeleteBuildpack(buildPlanBuildpack)).To(Succeed())
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})

	if revListOut.String() == "" {
		return "0.0.0", nil
	}

	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	err = gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--tags", strings.TrimSpace(revListOut.String())},
		Stdout: stdout,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(stdout.String(), "v")), nil
}
