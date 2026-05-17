package afk

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type PoolOpts struct {
	LogDir string
}

type Pool struct {
	maxConcurrency int
	opts           PoolOpts
	logger         *logger
	tasks          []func(ctx context.Context) error
	ids            []string
	mu             sync.Mutex
}

type TaskResult struct {
	ID       string
	Err      error
	Duration time.Duration
}

func NewPool(maxConcurrency int, opts PoolOpts) (*Pool, error) {
	l, err := newLogger(opts.LogDir)
	if err != nil {
		return nil, err
	}

	return &Pool{
		maxConcurrency: maxConcurrency,
		opts:           opts,
		logger:         l,
	}, nil
}

func (p *Pool) LogDir() string {
	return p.logger.runDir
}

func (p *Pool) Submit(id string, fn func(ctx context.Context) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tasks = append(p.tasks, fn)
	p.ids = append(p.ids, id)
}

func (p *Pool) Wait(ctx context.Context) []TaskResult {
	p.mu.Lock()
	tasks := p.tasks
	ids := p.ids
	p.tasks = nil
	p.ids = nil
	p.mu.Unlock()

	results := make([]TaskResult, len(tasks))
	sem := make(chan struct{}, p.maxConcurrency)
	var wg sync.WaitGroup

	for i, task := range tasks {
		if IsDraining(ctx) {
			results[i] = TaskResult{ID: ids[i], Err: ErrDraining}
			continue
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(idx int, fn func(ctx context.Context) error, id string) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			err := fn(ctx)
			results[idx] = TaskResult{
				ID:       id,
				Err:      err,
				Duration: time.Since(start),
			}
		}(i, task, ids[i])
	}

	wg.Wait()

	if err := p.logger.writeSummary(results); err != nil {
		fmt.Fprintf(os.Stderr, "afk: failed to write summary log: %v\n", err)
	}

	return results
}
