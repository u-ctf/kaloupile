package dependencies

import (
	"fmt"
	"os/exec"

	"github.com/yyewolf/kaloupile/pkg/config"
)

const (
	InfomaniakCredentialsTemplate = "cluster/dependencies/cert-manager/infomaniak-api-credentials.yaml"
	CertificateTemplate           = "cluster/dependencies/cert-manager/certificate.yaml"
)

func InstallCertManagerDependencies(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH: %w", err)
	}

	templates := []string{
		InfomaniakCredentialsTemplate,
		CertificateTemplate,
	}

	for _, templatePath := range templates {
		rendered, err := renderTemplate(templatePath, cfg)
		if err != nil {
			return err
		}

		if err := applyManifest(rendered); err != nil {
			return err
		}
	}

	return nil
}
