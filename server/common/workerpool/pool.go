// Package workerpool provides a bounded worker pool for async task execution.
// This prevents unbounded goroutine growth under high load.
package workerpool

import (
	"context"
	"sync"
)

// Task represents a unit of work to be executed
type Task func()

// Pool manages a fixed number of worker goroutines
type Pool struct {
	tasks   chan Task
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	workers int
}

// New creates a new worker pool with the specified number of workers and queue size
func New(workers, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		tasks:   make(chan Task, queueSize),
		ctx:     ctx,
		cancel:  cancel,
		workers: workers,
	}
	p.start()
	return p
}

// start spawns the worker goroutines
func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker processes tasks from the queue
func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task()
		}
	}
}

// Submit adds a task to the pool. Returns false if the queue is full.
func (p *Pool) Submit(task Task) bool {
	select {
	case p.tasks <- task:
		return true
	default:
		return false
	}
}

// SubmitWait adds a task to the pool, blocking if the queue is full.
// Returns false if the pool is shutting down.
func (p *Pool) SubmitWait(task Task) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.tasks <- task:
		return true
	}
}

// Stop gracefully shuts down the pool, waiting for all tasks to complete
func (p *Pool) Stop() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
}

// DefaultAuditPool is a shared pool for audit logging tasks
var DefaultAuditPool = New(4, 1000)

// DefaultEventPool is a shared pool for event publishing tasks
var DefaultEventPool = New(4, 1000)

// DefaultCachePool is a shared pool for cache write tasks
var DefaultCachePool = New(2, 500)
