package codex

import (
	"os/exec"

	"github.com/robinjoseph08/afk"
)

type Opts struct {
	Model    string
	FullAuto bool
}

type Agent struct {
	opts Opts
}

func New(opts Opts) *Agent {
	return &Agent{opts: opts}
}

func (a *Agent) Cmd(runOpts afk.RunOpts) *exec.Cmd {
	args := []string{}

	model := a.opts.Model
	if runOpts.Model != "" {
		model = runOpts.Model
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	if a.opts.FullAuto {
		args = append(args, "--full-auto")
	}

	args = append(args, runOpts.Prompt)

	return exec.Command("codex", args...)
}
