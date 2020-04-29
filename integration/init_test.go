package integration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/cloudfoundry/occam"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	mriBuildpack            string
	offlineMRIBuildpack     string
	bundlerBuildpack        string
	offlineBundlerBuildpack string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())

	bundlerBuildpack, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	offlineBundlerBuildpack, _, err = dagger.PackageCachedBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	mriBuildpack, err = dagger.GetLatestBuildpack("mri-cnb")
	Expect(err).ToNot(HaveOccurred())

	mriSource, err := dagger.GetLatestUnpackagedBuildpack("mri-cnb")
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
	}()

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}))
	suite("BuildpackYML", testBuildpackYML)
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
	buffer := bytes.NewBuffer(nil)
	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--abbrev=0", "--tags"},
		Stdout: buffer,
		Stderr: buffer,
	})
	if err != nil {
		if strings.Contains(buffer.String(), "No names found, cannot describe anything") {
			return "0.0.0", nil
		}

		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(buffer.String(), "v")), nil
}
