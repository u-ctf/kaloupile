package routes

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yyewolf/kaloupile/pkg/config"
)

const RoutesTemplatePath = "cluster/routes/routes.yaml"

func InstallRoutes(cfg *config.Config) error {
	return InstallRoutesFrom(cfg, RoutesTemplatePath)
}

func InstallRoutesFrom(cfg *config.Config, templatePath string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH: %w", err)
	}

	rendered, err := renderRoutesTemplate(templatePath, cfg)
	if err != nil {
		return err
	}

	_, err = runCommandStreaming("kubectl", "kubectl", bytes.NewReader(rendered), "apply", "-f", "-")
	return err
}

func renderRoutesTemplate(path string, data any) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve routes template path %s: %w", path, err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read routes template %s: %w", absPath, err)
	}

	tmpl, err := template.New(filepath.Base(absPath)).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse routes template %s: %w", absPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render routes template %s: %w", absPath, err)
	}

	return buf.Bytes(), nil
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
