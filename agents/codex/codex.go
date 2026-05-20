package codex

import (
	"os/exec"
	"strings"

	"github.com/robinjoseph08/afk"
)

type Opts struct {
	Model    string
	Effort   afk.Effort
	FullAuto bool
}

type Agent struct {
	opts Opts
}

func New(opts Opts) *Agent {
	return &Agent{opts: opts}
}

func (a *Agent) Cmd(runOpts afk.RunOpts) *exec.Cmd {
	args := []string{"exec"}

	model := a.opts.Model
	if runOpts.Model != "" {
		model = runOpts.Model
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	effort := a.opts.Effort
	if runOpts.Effort != 0 {
		effort = runOpts.Effort
	}
	if effortValue := effortString(effort); effortValue != "" {
		args = append(args, "--config", `model_reasoning_effort="`+effortValue+`"`)
	}

	if runOpts.Dir != "" {
		args = append(args, "--config", `projects."`+configKeySegment(runOpts.Dir)+`".trust_level="trusted"`)
	}

	if a.opts.FullAuto {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}

	args = append(args, runOpts.Prompt)

	return exec.Command("codex", args...)
}

func effortString(effort afk.Effort) string {
	switch effort {
	case afk.EffortLow:
		return "low"
	case afk.EffortMedium:
		return "medium"
	case afk.EffortHigh:
		return "high"
	default:
		return ""
	}
}

func configKeySegment(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
