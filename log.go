package afk

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultLogBase = ".afk/logs"

type logger struct {
	baseDir string
	runDir  string
}

func newLogger(logDir string) (*logger, error) {
	if logDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("afk: failed to get home dir: %w", err)
		}
		logDir = filepath.Join(home, defaultLogBase)
	}

	runDir := filepath.Join(logDir, time.Now().Format("2006-01-02T15-04-05"))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("afk: failed to create log directory: %w", err)
	}

	return &logger{baseDir: logDir, runDir: runDir}, nil
}

func (l *logger) logFile(id string) (*os.File, error) {
	name := fmt.Sprintf("%s.log", sanitizeFilename(id))
	path := filepath.Join(l.runDir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("afk: failed to create log file: %w", err)
	}
	return f, nil
}

func (l *logger) writeSummary(results []TaskResult) error {
	path := filepath.Join(l.runDir, "summary.log")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("afk: failed to create summary log: %w", err)
	}
	defer f.Close()

	for _, r := range results {
		status := "OK"
		if r.Err != nil {
			status = "FAIL"
		}
		fmt.Fprintf(f, "%s\t%s\t%s", status, r.ID, r.Duration)
		if r.Err != nil {
			fmt.Fprintf(f, "\t%v", r.Err)
		}
		fmt.Fprintln(f)
	}

	return nil
}

func sanitizeFilename(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '-')
		}
	}
	return string(result)
}
