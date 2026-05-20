package codex

import (
	"reflect"
	"testing"

	"github.com/robinjoseph08/afk"
)

func TestCmdUsesExecForNonInteractiveRuns(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5"})

	cmd := agent.Cmd(afk.RunOpts{Prompt: "fix the bug"})

	if cmd.Args[0] != "codex" {
		t.Fatalf("cmd.Args[0] = %q, want %q", cmd.Args[0], "codex")
	}
	want := []string{"codex", "exec", "--model", "gpt-5.5", "fix the bug"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCmdFullAutoBypassesInteractiveApprovals(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5", FullAuto: true})

	cmd := agent.Cmd(afk.RunOpts{Prompt: "fix the bug"})

	want := []string{
		"codex",
		"exec",
		"--model",
		"gpt-5.5",
		"--dangerously-bypass-approvals-and-sandbox",
		"fix the bug",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCmdRunOptsModelOverridesAgentModel(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5"})

	cmd := agent.Cmd(afk.RunOpts{Model: "gpt-5.4", Prompt: "fix the bug"})

	want := []string{"codex", "exec", "--model", "gpt-5.4", "fix the bug"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCmdConfiguresReasoningEffortFromAgentOpts(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5", Effort: afk.EffortHigh})

	cmd := agent.Cmd(afk.RunOpts{Prompt: "fix the bug"})

	want := []string{
		"codex",
		"exec",
		"--model",
		"gpt-5.5",
		"--config",
		`model_reasoning_effort="high"`,
		"fix the bug",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCmdRunOptsEffortOverridesAgentEffort(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5", Effort: afk.EffortHigh})

	cmd := agent.Cmd(afk.RunOpts{Effort: afk.EffortLow, Prompt: "fix the bug"})

	want := []string{
		"codex",
		"exec",
		"--model",
		"gpt-5.5",
		"--config",
		`model_reasoning_effort="low"`,
		"fix the bug",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCmdTrustsRunDirectory(t *testing.T) {
	agent := New(Opts{Model: "gpt-5.5"})

	cmd := agent.Cmd(afk.RunOpts{
		Dir:    "/Users/robinjoseph/.afk/worktrees/myrepo/agent-fix",
		Prompt: "fix the bug",
	})

	want := []string{
		"codex",
		"exec",
		"--model",
		"gpt-5.5",
		"--config",
		`projects."/Users/robinjoseph/.afk/worktrees/myrepo/agent-fix".trust_level="trusted"`,
		"fix the bug",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, want)
	}
}
