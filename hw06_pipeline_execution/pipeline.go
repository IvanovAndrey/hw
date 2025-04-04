package hw06pipelineexecution

import (
	"context"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil || len(stages) == 0 {
		return in
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-done
		cancel()
	}()
	ch := in
	for _, stage := range stages {
		ch = wrapStage(ctx, stage, ch)
	}

	out := make(Bi)
	go func() {
		defer close(out)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				out <- v
			}
		}
	}()
	return out
}

func wrapStage(ctx context.Context, stage Stage, in In) Out {
	out := make(Bi)
	go func() {
		stageOut := stage(in)
		defer func() {
			close(out)
			//nolint
			for range stageOut {
			}
		}()
		for v := range stageOut {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()
	return out
}
