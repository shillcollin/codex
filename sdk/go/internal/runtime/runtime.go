package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type LaunchConfig struct {
	CodexBin        string
	ConfigOverrides []string
	Cwd             string
	Env             map[string]string
}

func ResolveCodexBinary(path string) (string, error) {
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("codex binary not found at %s: %w", path, err)
		}
		return path, nil
	}

	resolved, err := exec.LookPath("codex")
	if err != nil {
		return "", errors.New("unable to locate codex binary; set Config.CodexBin or add codex to PATH")
	}
	return resolved, nil
}

func LaunchArgs(cfg LaunchConfig) []string {
	args := make([]string, 0, 2+len(cfg.ConfigOverrides)*2)
	for _, override := range cfg.ConfigOverrides {
		args = append(args, "--config", override)
	}
	args = append(args, "app-server", "--listen", "stdio://")
	return args
}

func LaunchEnv(overrides map[string]string) []string {
	env := os.Environ()
	if len(overrides) == 0 {
		return env
	}

	values := make(map[string]string, len(env)+len(overrides))
	for _, item := range env {
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				values[item[:i]] = item[i+1:]
				break
			}
		}
	}
	for key, value := range overrides {
		values[key] = value
	}

	merged := make([]string, 0, len(values))
	for key, value := range values {
		merged = append(merged, fmt.Sprintf("%s=%s", key, value))
	}
	return merged
}
