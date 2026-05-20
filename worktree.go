package afk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultWorktreeBase = ".afk/worktrees"

type WorktreeOpts struct {
	RepoDir string
	Branch  string
	BaseDir string   // defaults to ~/.afk/worktrees
	Setup   []string // command to run in the worktree after creation (e.g. []string{"make", "setup"})
}

type Worktree struct {
	Dir     string
	Branch  string
	repoDir string
}

func CreateWorktree(opts WorktreeOpts) (*Worktree, error) {
	baseDir := opts.BaseDir
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("afk: failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(home, defaultWorktreeBase)
	}

	repoName := filepath.Base(opts.RepoDir)
	branchDir := strings.ReplaceAll(opts.Branch, "/", "-")
	worktreeDir := filepath.Join(baseDir, repoName, branchDir)

	if _, err := os.Stat(worktreeDir); err == nil {
		return &Worktree{
			Dir:     worktreeDir,
			Branch:  opts.Branch,
			repoDir: opts.RepoDir,
		}, nil
	}

	if err := os.MkdirAll(filepath.Dir(worktreeDir), 0o755); err != nil {
		return nil, fmt.Errorf("afk: failed to create worktree directory: %w", err)
	}

	// Create the branch if it doesn't exist
	checkBranch := exec.Command("git", "-C", opts.RepoDir, "rev-parse", "--verify", opts.Branch)
	if err := checkBranch.Run(); err != nil {
		createBranch := exec.Command("git", "-C", opts.RepoDir, "branch", opts.Branch)
		if err := createBranch.Run(); err != nil {
			return nil, fmt.Errorf("afk: failed to create branch %s: %w", opts.Branch, err)
		}
	}

	cmd := exec.Command("git", "-C", opts.RepoDir, "worktree", "add", worktreeDir, opts.Branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("afk: failed to create worktree: %s: %w", string(out), err)
	}

	if len(opts.Setup) > 0 {
		setup := exec.Command(opts.Setup[0], opts.Setup[1:]...)
		setup.Dir = worktreeDir
		if out, err := setup.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("afk: worktree setup command failed: %s: %w", string(out), err)
		}
	}

	return &Worktree{
		Dir:     worktreeDir,
		Branch:  opts.Branch,
		repoDir: opts.RepoDir,
	}, nil
}

func (w *Worktree) Remove() error {
	// Remove the worktree
	cmd := exec.Command("git", "-C", w.repoDir, "worktree", "remove", w.Dir, "--force")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("afk: failed to remove worktree: %s: %w", string(out), err)
	}

	// Prune worktree metadata
	prune := exec.Command("git", "-C", w.repoDir, "worktree", "prune")
	_ = prune.Run()

	// Delete the branch
	delBranch := exec.Command("git", "-C", w.repoDir, "branch", "-D", w.Branch)
	if out, err := delBranch.CombinedOutput(); err != nil {
		return fmt.Errorf("afk: failed to delete branch %s: %s: %w", w.Branch, string(out), err)
	}

	// Clean up empty parent directories
	repoName := filepath.Base(w.repoDir)
	baseDir := filepath.Dir(filepath.Dir(w.Dir))
	repoWorktreeDir := filepath.Join(baseDir, repoName)
	entries, err := os.ReadDir(repoWorktreeDir)
	if err == nil && len(entries) == 0 {
		os.Remove(repoWorktreeDir)
	}

	return nil
}
