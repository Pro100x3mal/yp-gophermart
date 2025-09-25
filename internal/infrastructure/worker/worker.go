package worker

import (
	"sync"
)

type Task func()

type WorkerPool struct {
	numWorkers int
	queue      chan Task
	wg         sync.WaitGroup
}

func NewWorkerPool(limit int) *WorkerPool {
	numWorkers := limit
	if numWorkers <= 0 {
		numWorkers = 1
	}

	return &WorkerPool{
		numWorkers: numWorkers,
		queue:      make(chan Task, numWorkers),
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for task := range p.queue {
				task()
			}
		}()
	}
}

func (p *WorkerPool) Submit(t Task) {
	p.queue <- t
}

func (p *WorkerPool) Stop() {
	close(p.queue)
	p.wg.Wait()
}
