package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	clusterName = "kaloupile-dev"
	kindConfig  = "kind-config.yaml"
)

var projectRoot string

func main() {
	fmt.Println("🚀 U-CTF Local Development Setup")
	// Determine project root
	projectRoot = findProjectRoot()
	fmt.Printf("📁 Project root: %s\n", projectRoot)
	fmt.Println("=====================================")

	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		fmt.Printf("❌ Prerequisites check failed: %v\n", err)
		os.Exit(1)
	}

	// Check if cluster already exists
	if clusterExists() {
		fmt.Printf("ℹ️  Cluster '%s' already exists\n", clusterName)
		fmt.Print("Do you want to delete and recreate it? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			if err := deleteCluster(); err != nil {
				fmt.Printf("❌ Failed to delete cluster: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("✅ Using existing cluster")
			fmt.Println("\n📋 Next steps:")
			fmt.Println("   Run 'tilt up' to start the development environment")
			return
		}
	}

	// Create cluster
	if err := createCluster(); err != nil {
		fmt.Printf("❌ Failed to create cluster: %v\n", err)
		os.Exit(1)
	}

	// Set kubectl context
	if err := setKubectlContext(); err != nil {
		fmt.Printf("❌ Failed to set kubectl context: %v\n", err)
		os.Exit(1)
	}

	// Install prerequisites
	if err := installPrerequisites(); err != nil {
		fmt.Printf("❌ Failed to install prerequisites: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Setup complete!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("   1. Run 'tilt up' to start the development environment")
	fmt.Println("   2. Open http://localhost:10350 to view the Tilt dashboard")
	fmt.Println("\n📦 PostgreSQL will be available at:")
	fmt.Println("   Host: localhost")
	fmt.Println("   Port: 5432")
	fmt.Println("   Database: kaloupile")
	fmt.Println("   User: kaloupile")
	fmt.Println("   Password: kaloupile-dev-password")
}

func checkPrerequisites() error {
	fmt.Println("\n🔍 Checking prerequisites...")

	tools := []struct {
		name        string
		checkCmd    string
		installHint string
	}{
		{
			name:        "docker",
			checkCmd:    "docker version",
			installHint: "https://docs.docker.com/get-docker/",
		},
		{
			name:        "kind",
			checkCmd:    "kind version",
			installHint: "https://kind.sigs.k8s.io/docs/user/quick-start/#installation",
		},
		{
			name:        "kubectl",
			checkCmd:    "kubectl version --client",
			installHint: "https://kubernetes.io/docs/tasks/tools/",
		},
		{
			name:        "tilt",
			checkCmd:    "tilt version",
			installHint: "https://docs.tilt.dev/install.html",
		},
		{
			name:        "helm",
			checkCmd:    "helm version",
			installHint: "https://helm.sh/docs/intro/install/",
		},
	}

	allPresent := true
	for _, tool := range tools {
		if err := runSilent(tool.checkCmd); err != nil {
			fmt.Printf("   ❌ %s not found - install from: %s\n", tool.name, tool.installHint)
			allPresent = false
		} else {
			fmt.Printf("   ✅ %s\n", tool.name)
		}
	}

	if !allPresent {
		return fmt.Errorf("missing required tools")
	}

	// Check if Docker is running
	if err := runSilent("docker info"); err != nil {
		return fmt.Errorf("docker is not running - please start Docker")
	}

	return nil
}

func clusterExists() bool {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == clusterName {
			return true
		}
	}
	return false
}

func deleteCluster() error {
	fmt.Printf("\n🗑️  Deleting cluster '%s'...\n", clusterName)
	return runWithOutput("kind", "delete", "cluster", "--name", clusterName)
}

func createCluster() error {
	fmt.Printf("\n🔨 Creating KinD cluster '%s'...\n", clusterName)
	configPath := filepath.Join(projectRoot, kindConfig)
	return runWithOutput("kind", "create", "cluster", "--config", configPath)
}

func setKubectlContext() error {
	fmt.Println("\n⚙️  Setting kubectl context...")
	contextName := fmt.Sprintf("kind-%s", clusterName)
	return runWithOutput("kubectl", "config", "use-context", contextName)
}

func installPrerequisites() error {
	fmt.Println("\n📦 Installing prerequisites...")

	// 0. Create namespaces
	fmt.Println("\n   🔧 Creating namespaces...")
	if err := runWithOutput("kubectl", "apply", "-f", filepath.Join(projectRoot, "prerequisite", "namespaces", "namespaces.yaml")); err != nil {
		return fmt.Errorf("failed to create namespaces: %w", err)
	}

	// 1. Install Gateway API CRDs
	fmt.Println("\n   🔧 Installing Gateway API CRDs...")
	if err := runWithOutput("bash", filepath.Join(projectRoot, "prerequisite", "gateway-api", "install.sh")); err != nil {
		return fmt.Errorf("failed to install Gateway API CRDs: %w", err)
	}

	// 2. Install kGateway
	fmt.Println("\n   🔧 Installing kGateway...")
	if err := runWithOutput("bash", filepath.Join(projectRoot, "prerequisite", "kgateway", "install.sh")); err != nil {
		return fmt.Errorf("failed to install kGateway: %w", err)
	}

	// 3. Deploy Gateway
	fmt.Println("\n   🔧 Deploying Gateway...")
	if err := runWithOutput("kubectl", "apply", "-f", filepath.Join(projectRoot, "prerequisite", "gateway", "gateway.yaml")); err != nil {
		return fmt.Errorf("failed to deploy gateway: %w", err)
	}

	fmt.Println("\n   ✅ All prerequisites installed")
	return nil
}

func findProjectRoot() string {
	// First, check if we're running from the project directory
	if _, err := os.Stat(kindConfig); err == nil {
		cwd, _ := os.Getwd()
		return cwd
	}

	// Try to find go.mod to determine project root
	cwd, _ := os.Getwd()
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback to current directory
	return cwd
}

func runSilent(command string) error {
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func runWithOutput(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
