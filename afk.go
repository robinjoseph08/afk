package afk

import (
	"context"
	"os/exec"
	"time"
)

type Agent interface {
	Cmd(opts RunOpts) *exec.Cmd
}

type Effort int

const (
	EffortLow    Effort = 1
	EffortMedium Effort = 2
	EffortHigh   Effort = 3
)

type RunOpts struct {
	Dir        string
	Prompt     string
	PromptFile string
	OutputTag  string
	Model      string
	Effort     Effort
	Env        map[string]string
	LogFile    string // if set, agent output is written to this file
}

type RunResult[T any] struct {
	ExitCode int
	Output   T
	RawOutput string
	Dir      string
	Duration time.Duration
}

type NoOutput struct{}

type drainingKey struct{}

func IsDraining(ctx context.Context) bool {
	ch := ctx.Value(drainingKey{})
	if ch == nil {
		return false
	}
	select {
	case <-ch.(chan struct{}):
		return true
	default:
		return false
	}
}
