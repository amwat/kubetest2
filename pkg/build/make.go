package build

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

type MakeBuilder struct {
	RepoRoot string
}

var _ Builder = &MakeBuilder{}

const (
	target = "quick-release"
)

var (
	// This will need changed to support other platforms.
	tarballsToExtract = []string{
		"kubernetes.tar.gz",
		"kubernetes-test-linux-amd64.tar.gz",
		"kubernetes-test-portable.tar.gz",
		"kubernetes-client-linux-amd64.tar.gz",
	}
)

// Build builds kubernetes with the bazel-release make target
func (m *MakeBuilder) Build() (string, error) {
	version, err := sourceVersion(m.RepoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %v", err)
	}
	cmd := exec.Command("make", "-C", m.RepoRoot, target)
	exec.InheritOutput(cmd)
	if err = cmd.Run(); err != nil {
		return "", err
	}
	return version, extractBuiltTars(m.RepoRoot)
}

func extractBuiltTars(kuberoot string) error {
	allBuilds := filepath.Join(kuberoot, "_output", "gcs-stage")

	shouldExtract := make(map[string]bool)
	for _, name := range tarballsToExtract {
		shouldExtract[name] = true
	}

	err := filepath.Walk(allBuilds, func(path string, info os.FileInfo, err error) error { //Untar anything with the filename we want.
		if err != nil {
			return err
		}
		if shouldExtract[info.Name()] {
			klog.V(0).Infof("Extracting %s into current directory", path)
			//Extract it into current directory.
			cmd := exec.Command("tar", "-xzf", path)
			exec.InheritOutput(cmd)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("could not extract built tar archive: %v", err)
			}
			shouldExtract[info.Name()] = false
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not untar built archive: %v", err)
	}
	for n, s := range shouldExtract { // Make sure we found all the archives we were expecting.
		if s {
			return fmt.Errorf("expected built tarball was not present: %s", n)
		}
	}
	return nil
}
