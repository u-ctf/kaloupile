package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Domain   string         `yaml:"domain"`
	Scheme   string         `yaml:"scheme"`
	Postgres PostgresConfig `yaml:"postgres"`
	Kratos   KratosConfig   `yaml:"kratos"`
	Hydra    HydraConfig    `yaml:"hydra"`
}

type PostgresConfig struct {
	Host  string              `yaml:"host"`
	Port  int                 `yaml:"port"`
	Admin PostgresAdminConfig `yaml:"admin"`
	Users []PostgresUser      `yaml:"users"`
}

type PostgresAdminConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PostgresUser struct {
	Name      string   `yaml:"name"`
	Password  string   `yaml:"password"`
	Databases []string `yaml:"databases"`
}

// Helper to get a user by name
func (c *Config) GetPostgresUser(name string) *PostgresUser {
	for _, u := range c.Postgres.Users {
		if u.Name == name {
			return &u
		}
	}
	return nil
}

type KratosConfig struct {
	PublicPort int                  `yaml:"public_port"`
	AdminPort  int                  `yaml:"admin_port"`
	Postgres   KratosPostgresConfig `yaml:"postgres"`
	GitHub     KratosGitHubConfig   `yaml:"github"`
	Secrets    KratosSecretsConfig  `yaml:"secrets"`
	Webhook    KratosWebhookConfig  `yaml:"webhook"`
}

type KratosPostgresConfig struct {
	User     string `yaml:"user"`
	Database string `yaml:"database"`
}

type KratosGitHubConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type KratosSecretsConfig struct {
	Cookie string `yaml:"cookie"`
	Cipher string `yaml:"cipher"`
}

type KratosWebhookConfig struct {
	LocalIP string `yaml:"local_ip"`
	Port    int    `yaml:"port"`
}

type HydraConfig struct {
	PublicPort int                 `yaml:"public_port"`
	AdminPort  int                 `yaml:"admin_port"`
	Postgres   HydraPostgresConfig `yaml:"postgres"`
	Secrets    HydraSecretsConfig  `yaml:"secrets"`
}

type HydraPostgresConfig struct {
	User     string `yaml:"user"`
	Database string `yaml:"database"`
}

type HydraSecretsConfig struct {
	System string `yaml:"system"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <project-root>")
		os.Exit(1)
	}

	projectRoot := os.Args[1]

	// Load config
	configPath := filepath.Join(projectRoot, "config.yml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ Failed to read config.yml: %v\n", err)
		os.Exit(1)
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		fmt.Printf("❌ Failed to parse config.yml: %v\n", err)
		os.Exit(1)
	}

	// Generate Kratos ConfigMap
	if err := generateFromTemplate(
		filepath.Join(projectRoot, "external", "kratos", "kratos-config.yaml.tmpl"),
		filepath.Join(projectRoot, "external", "kratos", "kratos-config.generated.yaml"),
		config,
	); err != nil {
		fmt.Printf("❌ Failed to generate kratos-config.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated external/kratos/kratos-config.generated.yaml")

	// Generate Hydra ConfigMap
	if err := generateFromTemplate(
		filepath.Join(projectRoot, "external", "hydra", "hydra-config.yaml.tmpl"),
		filepath.Join(projectRoot, "external", "hydra", "hydra-config.generated.yaml"),
		config,
	); err != nil {
		fmt.Printf("❌ Failed to generate hydra-config.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated external/hydra/hydra-config.generated.yaml")

	// Generate PostgreSQL sync script
	if err := generateFromTemplate(
		filepath.Join(projectRoot, "external", "postgresql", "sync.sh.tmpl"),
		filepath.Join(projectRoot, "external", "postgresql", "sync.generated.sh"),
		config,
	); err != nil {
		fmt.Printf("❌ Failed to generate postgres sync script: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated external/postgresql/sync.generated.sh")

	// CHMOD the sync script to be executable
	syncScriptPath := filepath.Join(projectRoot, "external", "postgresql", "sync.generated.sh")
	if err := os.Chmod(syncScriptPath, 0755); err != nil {
		fmt.Printf("❌ Failed to chmod postgres sync script: %v\n", err)
		os.Exit(1)
	}

	// Generate HTTPRoutes
	if err := generateFromTemplate(
		filepath.Join(projectRoot, "routes", "routes.yaml.tmpl"),
		filepath.Join(projectRoot, "routes", "routes.generated.yaml"),
		config,
	); err != nil {
		fmt.Printf("❌ Failed to generate routes: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated routes/routes.generated.yaml")

	// Generate Mailslurper config
	if err := generateFromTemplate(
		filepath.Join(projectRoot, "external", "mailslurper", "mailslurper-config.yaml.tmpl"),
		filepath.Join(projectRoot, "external", "mailslurper", "mailslurper-config.generated.yaml"),
		config,
	); err != nil {
		fmt.Printf("❌ Failed to generate mailslurper config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated external/mailslurper/mailslurper-config.generated.yaml")

	fmt.Println("\n🎉 Configuration generated successfully!")
}

func generateFromTemplate(templatePath, outputPath string, config Config) error {
	tmplData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Create template with helper functions
	funcMap := template.FuncMap{
		"getUserPassword": func(users []PostgresUser, name string) string {
			for _, u := range users {
				if u.Name == name {
					return u.Password
				}
			}
			return ""
		},
	}

	tmpl, err := template.New("config").Funcs(funcMap).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
