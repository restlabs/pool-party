package pool_party

import (
	"github.com/we-are-discussing-rest/pool-party/internal"
	"sync"
	"sync/atomic"
)

// Pool represents a concurrent worker pool for executing tasks.
//
// The pool manages a fixed number of workers that concurrently process tasks
// submitted to the task queue. Each worker is represented by an internal.Worker
// instance. The pool provides a mechanism to submit tasks to the workers and
// gracefully shut down the pool when no more tasks are expected.
//
// Fields:
//   - workers: A slice of internal.Worker instances representing the worker pool.
//   - count: The total number of workers in the pool.
//   - taskId: An atomic counter to generate unique task IDs.
//   - taskQueue: A channel for submitting tasks to be processed by the workers.
//   - workerWg: A WaitGroup to track the active workers in the pool.
//   - shutdown: A channel used to signal the pool to shut down gracefully.
//   - ready: A WaitGroup to track the readiness of the workers.
type Pool struct {
	workers   []*internal.Worker
	count     int
	taskId    atomic.Uint64
	taskQueue chan internal.Task
	workerWg  sync.WaitGroup
	shutdown  chan bool
	ready     sync.WaitGroup
}

// NewPool creates a new concurrent worker pool with the specified number of workers.
//
// The function initializes a Pool structure with the given count of workers. It
// creates a task queue channel for submitting tasks to the workers and sets up
// the necessary synchronization primitives for managing the pool's lifecycle.
// Each worker is assigned a unique ID and associated with the task queue channel.
//
// Parameters:
//   - count: The number of workers in the pool.
//
// Returns:
//
//	A pointer to the newly created Pool.
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

// Start initiates the concurrent execution of tasks by activating all workers in the pool.
//
// The method increments the workerWg WaitGroup for each active worker, signaling
// their readiness to process tasks. It then launches a goroutine to wait for all
// workers to complete their tasks before closing the task queue channel.
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

// Submit adds a new task to the pool for concurrent execution by available workers.
//
// The method takes a task function and wraps it in an internal.Task structure with
// a unique ID generated from the pool's atomic counter. The task is then submitted
// to the task queue channel for processing by available workers.
//
// If the pool is in the process of shutting down (signaled by the closed shutdown
// channel), the task submission is ignored.
//
// Parameters:
//   - task: The function representing the task to be executed by a worker.
func (p *Pool) Submit(task func()) {
	// Create a new task with a unique ID and the provided task function.
	t := internal.Task{
		ID:  p.taskId.Add(1),
		Job: task,
	}

	// Attempt to submit the task to the task queue unless the pool is in the process
	// of shutting down (signaled by the closed shutdown channel).
	select {
	case <-p.shutdown:
		return
	default:
		p.taskQueue <- t
	}
}

// Stop initiates the graceful shutdown of the worker pool.
//
// The method waits for all workers to become ready before closing the shutdown
// channel, signaling to workers that no more tasks will be submitted. Afterward,
// it calls the Stop method for each worker to ensure they complete any ongoing tasks
// and terminate gracefully.
//
// Note: The pool assumes that all tasks have been submitted and that the workers
// are ready before calling Stop.
func (p *Pool) Stop() {
	// Wait for all workers to be in ready state before closing the shutdown channel.
	p.ready.Wait()
	// Close the shutdown channel to signal to workers that no more tasks will be submitted.
	close(p.shutdown)
	// Stop each worker to ensure they complete ongoing tasks and terminate gracefully.
	for _, worker := range p.workers {
		worker.Stop()
	}
}
