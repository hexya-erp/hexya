// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"sync"
	"time"
)

// A WorkerFunction can be executed in a loop in background every given LoopPeriod.
type WorkerFunction interface {
	// Run the worker function
	Run()
	// LoopPeriod is the time between each run of the worker function
	LoopPeriod() time.Duration
}

// A workerFunction implements WorkerFunction
type workerFunction struct {
	fnct   func()
	period time.Duration
}

// Run the workerFunction
func (w *workerFunction) Run() {
	w.fnct()
}

// LoopPeriod is the time between each run of the worker function
func (w *workerFunction) LoopPeriod() time.Duration {
	return w.period
}

// NewWorkerFunction returns a WorkerFunction from the given fnct and period
func NewWorkerFunction(fnct func(), period time.Duration) WorkerFunction {
	return &workerFunction{
		fnct:   fnct,
		period: period,
	}
}

var (
	workerFunctions []WorkerFunction
	workerStop      chan struct{}
	workerGroup     sync.WaitGroup
)

// RegisterWorker registers a WorkerFunction so that it will be called by the core loop.
func RegisterWorker(wf WorkerFunction) {
	workerFunctions = append(workerFunctions, wf)
}

// RunWorkerLoop launches the hexya core worker loop.
//
// This function must be called only once or it will panic
func RunWorkerLoop() {
	if workerStop != nil {
		log.Panic("RunWorkerLoop must be called only once.")
	}
	workerStop = make(chan struct{})
	for _, workerFunc := range workerFunctions {
		workerGroup.Add(1)
		go func() {
			ticker := time.NewTicker(workerFunc.LoopPeriod())
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					workerFunc.Run()
				case <-workerStop:
					workerGroup.Done()
					return
				}
			}
		}()
	}
}

// StopWorkerLoop stops the hexya core worker loop.
//
// Calling this method if the core worker loop is not running will cause panic.
func StopWorkerLoop() {
	close(workerStop)
	workerGroup.Wait()
	workerStop = nil
}
