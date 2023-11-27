package internal

import (
	"fmt"
	"sync"
)

type Task struct {
	ID  uint64
	Job func()
}

type Worker struct {
	ID   uint64
	Task chan Task
	Quit chan bool
}

func NewWorker(id uint64, task chan Task) *Worker {
	return &Worker{
		ID:   id,
		Task: task,
		Quit: make(chan bool),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup, swg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		fmt.Printf("starting worker %d\n", w.ID)

		for {
			select {
			case task, ok := <-w.Task:
				swg.Add(1)
				if !ok {
					swg.Done()
					return
				}

				fmt.Printf("Worker ID %d, execution Task ID %d\n", w.ID, task.ID)
				task.Job()

				swg.Done()
			case <-w.Quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
