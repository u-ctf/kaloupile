package dependencies

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

const FakeSMTPManifestPath = "cluster/dependencies/fake-smtp/fake-smtp.yaml"

func InstallFakeSMTP() error {
	return InstallFakeSMTPFrom(FakeSMTPManifestPath)
}

func InstallFakeSMTPFrom(path string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve fake smtp manifest path %s: %w", path, err)
	}

	_, err = runKubectlStreaming("apply", "-f", absPath)
	return err
}
