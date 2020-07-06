package progress

type WorkerPool struct {
	jobs     chan int
	start    chan int
	numJobs  int
	finished chan int
	workers  int
}

func NewWorkerPool(workers int) *WorkerPool {
	w := &WorkerPool{
		jobs:     make(chan int),
		start:    make(chan int),
		numJobs:  0,
		finished: make(chan int),
		workers:  workers,
	}

	for i := 1; i <= workers; i++ {
		go w.worker(i)
	}

	return w
}

func (w *WorkerPool) worker(id int) {
	for range w.jobs {
		w.start <- 1
		<-w.finished
	}
}

func (w *WorkerPool) Add() {
	w.jobs <- 1
	<-w.start
}

func (w *WorkerPool) Close() {
	close(w.jobs)
}

func (w *WorkerPool) Done() {
	w.finished <- 1
}
