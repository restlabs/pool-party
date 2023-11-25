package main

import (
	"github.com/we-are-discussing-rest/pool-party/internal"
	"sync"
	"sync/atomic"
)

type Pool struct {
	workers   []*internal.Worker
	count     int
	taskId    atomic.Uint64
	taskQueue chan internal.Task
	workerWg  sync.WaitGroup
	shutdown  chan bool
	ready     sync.WaitGroup
}

func NewPool(count int) *Pool {
	p := &Pool{
		workers:   make([]*internal.Worker, count),
		count:     count,
		taskQueue: make(chan internal.Task),
		shutdown:  make(chan bool),
		ready:     sync.WaitGroup{},
	}
	for i := 0; i < count; i++ {
		p.workers[i] = internal.NewWorker(uint64(i+1), p.taskQueue)
	}

	return p
}

func (p *Pool) Start() {
	for _, w := range p.workers {
		p.workerWg.Add(1)
		w.Start(&p.workerWg, &p.ready)
	}

	go func() {
		p.workerWg.Wait()
		close(p.taskQueue)
	}()
}

func (p *Pool) Submit(task func()) {
	t := internal.Task{
		ID:  p.taskId.Add(1),
		Job: task,
	}
	select {
	case <-p.shutdown:
		return
	default:
		p.taskQueue <- t
	}
}

func (p *Pool) Stop() {
	p.ready.Wait()
	close(p.shutdown)

	for _, worker := range p.workers {
		worker.Stop()
	}
}
