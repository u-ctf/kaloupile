package dependencies

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yyewolf/kaloupile/pkg/config"
)

const (
	PostgresNamespace            = "external"
	PostgresTemplatePath         = "cluster/dependencies/postgresql/postgresql.yaml"
	PostgresConfigHashAnnotation = "kaloupile.dev/postgresql-config-hash"
)

func InstallPostgreSQL(cfg *config.Config) error {
	return InstallPostgreSQLFrom(cfg, PostgresTemplatePath, true)
}

func InstallPostgreSQLFrom(cfg *config.Config, templatePath string, deleteBeforeApply bool) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH: %w", err)
	}

	hash, err := postgresConfigHash(cfg)
	if err != nil {
		return err
	}

	currentHash, exists, err := getNamespaceAnnotation(PostgresNamespace, PostgresConfigHashAnnotation)
	if err != nil {
		return err
	}

	rendered, err := renderTemplate(templatePath, cfg)
	if err != nil {
		return err
	}

	if exists && currentHash != hash && deleteBeforeApply {
		if err := deleteManifest(rendered); err != nil {
			return err
		}
	}

	if err := applyManifest(rendered); err != nil {
		return err
	}

	if err := annotateNamespace(PostgresNamespace, PostgresConfigHashAnnotation, hash); err != nil {
		return err
	}

	return nil
}

func postgresConfigHash(cfg *config.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}

	admin := cfg.Postgres.Admin
	input := fmt.Sprintf("user=%s\npassword=%s\ndatabase=%s\n", admin.User, admin.Password, admin.Database)
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:]), nil
}

func renderTemplate(path string, data any) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve template path %s: %w", path, err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", absPath, err)
	}

	tmpl, err := template.New(filepath.Base(absPath)).Funcs(template.FuncMap{
		"b64enc": func(input string) string {
			return base64.StdEncoding.EncodeToString([]byte(input))
		},
	}).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", absPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template %s: %w", absPath, err)
	}

	return buf.Bytes(), nil
}

func getNamespaceAnnotation(namespace, key string) (string, bool, error) {
	output, err := runKubectl("get", "namespace", namespace, "-o", "jsonpath={.metadata.annotations}")
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found") {
			return "", false, nil
		}
		return "", false, err
	}

	annotations := parseAnnotations(output)
	value, exists := annotations[key]

	return strings.TrimSpace(value), exists, nil
}

func parseAnnotations(input string) map[string]string {
	var annotations map[string]string
	json.Unmarshal([]byte(input), &annotations)
	return annotations
}

func deleteManifest(content []byte) error {
	_, err := runKubectlStreamingWithStdin(bytes.NewReader(content), "delete", "-f", "-", "--wait=true", "--ignore-not-found")
	return err
}

func applyManifest(content []byte) error {
	_, err := runKubectlStreamingWithStdin(bytes.NewReader(content), "apply", "-f", "-")
	return err
}

func annotateNamespace(namespace, key, value string) error {
	annotation := fmt.Sprintf("%s=%s", key, value)
	_, err := runKubectlStreaming("annotate", "namespace", namespace, annotation, "--overwrite")
	return err
}

func runKubectl(args ...string) (string, error) {
	return runCommandCapture("kubectl", args...)
}

func runKubectlStreaming(args ...string) (string, error) {
	return runCommandStreaming("kubectl", "kubectl", nil, args...)
}

func runKubectlStreamingWithStdin(stdin io.Reader, args ...string) (string, error) {
	return runCommandStreaming("kubectl", "kubectl", stdin, args...)
}

func runCommandCapture(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w: %s", bin, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func runCommandStreaming(prefix, bin string, stdin io.Reader, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	var output bytes.Buffer
	stdoutWriter := newPrefixWriter(prefix, os.Stdout)
	stderrWriter := newPrefixWriter(prefix, os.Stderr)
	cmd.Stdout = io.MultiWriter(stdoutWriter, &output)
	cmd.Stderr = io.MultiWriter(stderrWriter, &output)

	err := cmd.Run()
	_ = stdoutWriter.Flush()
	_ = stderrWriter.Flush()
	if err != nil {
		return output.String(), fmt.Errorf("%s %s failed: %w: %s", bin, strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}

	return output.String(), nil
}

type prefixWriter struct {
	dst     io.Writer
	prefix  string
	pending string
}

func newPrefixWriter(prefix string, dst io.Writer) *prefixWriter {
	return &prefixWriter{dst: dst, prefix: prefix}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	chunk := w.pending + string(p)
	lines := strings.Split(chunk, "\n")
	for i := 0; i < len(lines)-1; i++ {
		if _, err := fmt.Fprintf(w.dst, "[%s] %s\n", w.prefix, lines[i]); err != nil {
			return len(p), err
		}
	}
	w.pending = lines[len(lines)-1]
	return len(p), nil
}

func (w *prefixWriter) Flush() error {
	if w.pending == "" {
		return nil
	}
	_, err := fmt.Fprintf(w.dst, "[%s] %s\n", w.prefix, w.pending)
	w.pending = ""
	return err
}
