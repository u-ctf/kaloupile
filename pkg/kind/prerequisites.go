package kind

import (
	"fmt"
	"path/filepath"
)

const DefaultPrerequisitesScript = "cluster/prerequisites/install-prerequisites.sh"

func InstallPrerequisites() error {
	return InstallPrerequisitesFrom(DefaultPrerequisitesScript)
}

func InstallPrerequisitesFrom(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve prerequisites path %s: %w", path, err)
	}

	_, err = runCommandStreaming("prerequisites", "bash", absPath)
	return err
}
