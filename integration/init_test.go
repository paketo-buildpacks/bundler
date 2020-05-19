package integration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/paketo-buildpacks/occam"
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

	defer func() {
		dagger.DeleteBuildpack(mriBuildpack)
		dagger.DeleteBuildpack(offlineMRIBuildpack)
		dagger.DeleteBuildpack(bundlerBuildpack)
		dagger.DeleteBuildpack(offlineBundlerBuildpack)
		dagger.DeleteBuildpack(buildPlanBuildpack)
	}()

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}))
	suite("buildpack.yml", testBuildpackYML)
	suite("gemfile.lock", testGemfileLock)
	suite("Default", testDefault)
	suite("Logging", testLogging)
	suite("Offline", testOffline)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)
	suite.Run(t)
}

func ContainerLogs(id string) func() string {
	docker := occam.NewDocker()

	return func() string {
		logs, _ := docker.Container.Logs.Execute(id)
		return logs.String()
	}
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})
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
