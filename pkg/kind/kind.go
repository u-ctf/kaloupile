package kind

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ClusterName       = "kaloupile-dev"
	ClusterConfigPath = "kind-config.yml"
)

func EnsureCluster() error {
	if _, err := exec.LookPath("kind"); err != nil {
		return fmt.Errorf("kind not found in PATH: %w", err)
	}

	clusters, err := listClusters()
	if err != nil {
		return err
	}

	hasTarget := false
	for _, name := range clusters {
		if name == ClusterName {
			hasTarget = true
			continue
		}
		if err := deleteCluster(name); err != nil {
			return fmt.Errorf("delete kind cluster %s: %w", name, err)
		}
	}

	if !hasTarget {
		if err := createCluster(ClusterName); err != nil {
			return err
		}
	}

	return nil
}

func listClusters() ([]string, error) {
	output, err := runKind("get", "clusters")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	clusters := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		clusters = append(clusters, name)
	}

	return clusters, nil
}

func createCluster(name string) error {
	configPath, err := filepath.Abs(ClusterConfigPath)
	if err != nil {
		return fmt.Errorf("resolve kind config path %s: %w", ClusterConfigPath, err)
	}

	_, err = runKindLogged("create", "cluster", "--name", name, "--config", configPath)
	if err != nil {
		return fmt.Errorf("create kind cluster %s: %w", name, err)
	}

	return nil
}

func runKind(args ...string) (string, error) {
	return runCommandCapture("kind", args...)
}

func runKindLogged(args ...string) (string, error) {
	return runCommandStreaming("kind", "kind", args...)
}

func runCommandCapture(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w: %s", bin, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func runCommandStreaming(prefix, bin string, args ...string) (string, error) {
	return runCommandStreamingWithStdin(prefix, bin, nil, args...)
}

func runCommandStreamingWithStdin(prefix, bin string, stdin io.Reader, args ...string) (string, error) {
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
