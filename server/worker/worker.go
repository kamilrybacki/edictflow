package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

type Config struct {
	CheckInterval time.Duration
}

type ChangeHandler interface {
	HandleExpiredTemporary(ctx context.Context) error
}

type Worker struct {
	config        Config
	changeHandler ChangeHandler
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func New(config Config, changeHandler ChangeHandler) *Worker {
	if config.CheckInterval == 0 {
		config.CheckInterval = 60 * time.Second
	}

	return &Worker{
		config:        config,
		changeHandler: changeHandler,
		stopCh:        make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
	log.Printf("worker started with check interval %v", w.config.CheckInterval)
}

func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	log.Printf("worker stopped")
}

func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.CheckInterval)
	defer ticker.Stop()

	// Run immediately on start
	w.checkExpiredChanges(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.checkExpiredChanges(ctx)
		}
	}
}

func (w *Worker) checkExpiredChanges(ctx context.Context) {
	if w.changeHandler == nil {
		return
	}

	if err := w.changeHandler.HandleExpiredTemporary(ctx); err != nil {
		log.Printf("worker: failed to handle expired changes: %v", err)
	}
}
