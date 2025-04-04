package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrWorkersNumInvalid   = errors.New("number of workers must be greater than 0")
	ErrErrorLimitInvalid   = errors.New("error limit must be greater than 0")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrWorkersNumInvalid
	}
	if m <= 0 {
		return ErrErrorLimitInvalid
	}

	taskChan := make(chan Task)
	errChan := make(chan error, n)
	var errCounter int32
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := func(ctx context.Context) {
		defer wg.Done()
		for task := range taskChan {
			if int(atomic.LoadInt32(&errCounter)) >= m {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				if err := task(); err != nil {
					errChan <- err
				}
			}
		}
	}
	go errCollector(cancel, errChan, &errCounter, m)
	go taskProducer(ctx, tasks, &errCounter, m, taskChan)
	startWorkers(ctx, worker, &wg, n)

	wg.Wait()
	close(errChan)
	if int(atomic.LoadInt32(&errCounter)) >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func taskProducer(
	ctx context.Context,
	tasks []Task,
	errCounter *int32,
	m int,
	taskChan chan<- Task,
) {
	defer close(taskChan)
	for _, task := range tasks {
		if int(atomic.LoadInt32(errCounter)) >= m {
			return
		}
		select {
		case <-ctx.Done():
			return
		case taskChan <- task:
		}
	}
}

func errCollector(cancel context.CancelFunc, errChan <-chan error, errCounter *int32, m int) {
	for err := range errChan {
		if err != nil {
			atomic.AddInt32(errCounter, 1)
			if int(atomic.LoadInt32(errCounter)) >= m {
				cancel()
				return
			}
		}
	}
}

func startWorkers(ctx context.Context, worker func(context.Context), wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(ctx)
	}
}
