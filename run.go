package afk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"
)

func Run[T any](ctx context.Context, agent Agent, opts RunOpts) (*RunResult[T], error) {
	if IsDraining(ctx) {
		return nil, ErrDraining
	}

	prompt, err := resolvePrompt(opts)
	if err != nil {
		return nil, err
	}
	opts.Prompt = prompt

	cmd := agent.Cmd(opts)
	cmd.Dir = opts.Dir

	cmd.Env = os.Environ()
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var outputBuf bytes.Buffer
	var writer io.Writer = &outputBuf

	if opts.LogFile != "" {
		logFile, err := os.Create(opts.LogFile)
		if err != nil {
			return nil, fmt.Errorf("afk: failed to create log file: %w", err)
		}
		defer logFile.Close()
		writer = io.MultiWriter(&outputBuf, logFile)
	}

	cmd.Stdout = writer
	cmd.Stderr = writer

	start := time.Now()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("afk: failed to start agent: %w", err)
	}

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	var cmdErr error
	select {
	case <-ctx.Done():
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case cmdErr = <-doneCh:
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			cmdErr = <-doneCh
		}
	case cmdErr = <-doneCh:
	}

	duration := time.Since(start)
	rawOutput := outputBuf.String()

	exitCode := 0
	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("afk: agent process error: %w", cmdErr)
		}
	}

	result := &RunResult[T]{
		ExitCode:  exitCode,
		RawOutput: rawOutput,
		Dir:       opts.Dir,
		Duration:  duration,
	}

	var zero T
	if _, isNoOutput := any(zero).(NoOutput); !isNoOutput {
		parsed, err := parseOutput[T](rawOutput, opts.OutputTag)
		if err != nil {
			return result, err
		}
		result.Output = parsed
	}

	return result, nil
}

func resolvePrompt(opts RunOpts) (string, error) {
	if opts.Prompt != "" && opts.PromptFile != "" {
		return "", ErrPromptConflict
	}
	if opts.Prompt == "" && opts.PromptFile == "" {
		return "", ErrPromptEmpty
	}
	if opts.PromptFile != "" {
		data, err := os.ReadFile(opts.PromptFile)
		if err != nil {
			return "", fmt.Errorf("afk: failed to read prompt file: %w", err)
		}
		return string(data), nil
	}
	return opts.Prompt, nil
}

func parseOutput[T any](raw string, tag string) (T, error) {
	var zero T
	if tag == "" {
		tag = "result"
	}

	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>([\s\S]*?)</%s>`, regexp.QuoteMeta(tag), regexp.QuoteMeta(tag)))
	matches := pattern.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return zero, ErrOutputMissing
	}

	// Use the last match
	jsonStr := matches[len(matches)-1][1]

	var result T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return zero, fmt.Errorf("%w: %v", ErrOutputParse, err)
	}

	return result, nil
}
