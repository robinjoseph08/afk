package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robinjoseph08/afk"
	"github.com/robinjoseph08/afk/agents/codex"
)

type ReviewResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

func main() {
	ctx := afk.SignalContext()

	implementer := codex.New(codex.Opts{Model: "o3", FullAuto: true})
	reviewer := codex.New(codex.Opts{Model: "o4-mini", FullAuto: true})

	pool, err := afk.NewPool(3, afk.PoolOpts{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}

	tasks := []struct {
		id      string
		repoDir string
		prompt  string
	}{
		{"fix-auth", "/home/user/code/myapp", "fix the authentication bug in login.go"},
		{"add-tests", "/home/user/code/mylib", "add unit tests for the parser package"},
	}

	for _, task := range tasks {
		task := task
		pool.Submit(task.id, func(ctx context.Context) error {
			wt, err := afk.CreateWorktree(afk.WorktreeOpts{
				RepoDir: task.repoDir,
				Branch:  "agent/" + task.id,
			})
			if err != nil {
				return err
			}
			defer wt.Remove()

			// Implement
			_, err = afk.Run[afk.NoOutput](ctx, implementer, afk.RunOpts{
				Dir:     wt.Dir,
				Prompt:  task.prompt,
				LogFile: filepath.Join(pool.LogDir(), task.id+"-implement.log"),
			})
			if errors.Is(err, afk.ErrDraining) {
				return nil
			}
			if err != nil {
				return err
			}

			// Review
			review, err := afk.Run[ReviewResult](ctx, reviewer, afk.RunOpts{
				Dir:       wt.Dir,
				Prompt:    "Review the changes. Output <review>{\"approved\": bool, \"feedback\": \"...\"}</review>",
				OutputTag: "review",
				LogFile:   filepath.Join(pool.LogDir(), task.id+"-review.log"),
			})
			if errors.Is(err, afk.ErrDraining) {
				return nil
			}
			if err != nil {
				return err
			}

			if !review.Output.Approved {
				// Fix based on feedback
				_, err = afk.Run[afk.NoOutput](ctx, implementer, afk.RunOpts{
					Dir:     wt.Dir,
					Prompt:  fmt.Sprintf("Address this review feedback: %s", review.Output.Feedback),
					LogFile: filepath.Join(pool.LogDir(), task.id+"-fix.log"),
				})
				if err != nil && !errors.Is(err, afk.ErrDraining) {
					return err
				}
			}

			return nil
		})
	}

	results := pool.Wait(ctx)
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("FAIL %s (%s): %v\n", r.ID, r.Duration, r.Err)
		} else {
			fmt.Printf("OK   %s (%s)\n", r.ID, r.Duration)
		}
	}
}
