package stdio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/openai/codex/sdk/go/internal/runtime"
)

type Config struct {
	CodexBin        string
	ConfigOverrides []string
	Cwd             string
	Env             map[string]string
}

type Transport struct {
	cmd *exec.Cmd

	stdin  io.WriteCloser
	stdout *bufio.Reader

	writeMu sync.Mutex

	stderrMu    sync.Mutex
	stderrLines []string
}

func Start(ctx context.Context, cfg Config) (*Transport, error) {
	bin, err := runtime.ResolveCodexBinary(cfg.CodexBin)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, bin, runtime.LaunchArgs(runtime.LaunchConfig{
		CodexBin:        cfg.CodexBin,
		ConfigOverrides: cfg.ConfigOverrides,
		Cwd:             cfg.Cwd,
		Env:             cfg.Env,
	})...)
	cmd.Dir = cfg.Cwd
	cmd.Env = runtime.LaunchEnv(cfg.Env)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	t := &Transport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}
	go t.drainStderr(stderr)
	return t, nil
}

func (t *Transport) ReadLine() ([]byte, error) {
	line, err := t.stdout.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("app-server closed stdout: %s", t.StderrTail())
		}
		return nil, err
	}
	return bytes.TrimRight(line, "\r\n"), nil
}

func (t *Transport) WriteJSON(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func (t *Transport) Close() error {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.cmd == nil || t.cmd.Process == nil {
		return nil
	}
	if err := t.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	_, _ = t.cmd.Process.Wait()
	return nil
}

func (t *Transport) StderrTail() string {
	t.stderrMu.Lock()
	defer t.stderrMu.Unlock()
	return strings.Join(t.stderrLines, "\n")
}

func (t *Transport) drainStderr(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t.stderrMu.Lock()
		t.stderrLines = append(t.stderrLines, scanner.Text())
		if len(t.stderrLines) > 40 {
			t.stderrLines = append([]string(nil), t.stderrLines[len(t.stderrLines)-40:]...)
		}
		t.stderrMu.Unlock()
	}
}
