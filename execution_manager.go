// Copyright (C) 2022 Blinnikov AA <goofinator@mail.ru>
//
// This file is part of execution.
//
// execution is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// execution is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with execution.  If not, see <http://www.gnu.org/licenses/>.

package execution

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ExecutionManagerConfig is a configuration of ExecutionManager
type ExecutionManagerConfig struct {
	// Signals is a list of syscall.Signals capable to initiate shutdown (default: [SIGINT, SIGTERM])
	Signals []os.Signal
	// Log is a Logger that will be used to display a processes execution status (default: nil - no output)
	Log Logger
}

// NewDefaultExecutionManager creates ExecutionManager
// with default config
func NewDefaultExecutionManager() *ExecutionManager {
	config := ExecutionManagerConfig{}
	return NewExecutionManager(config)
}

// NewExecutionManager creates ExecutionManager with given configuration
func NewExecutionManager(config ExecutionManagerConfig) *ExecutionManager {
	if config.Signals == nil {
		config.Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, config.Signals...)
	return &ExecutionManager{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		wg:         &sync.WaitGroup{},
		sigChan:    signChan,
		log:        config.Log,
	}
}

// The ExecutionManager runs multiple processes in a controlled manner.
// A graceful termination will be initiated if one of the conditions is met:
//  - Programm receives one of the configured signals.
//  - One of the processes returns from it's Run method
type (
	ExecutionManager struct {
		ctx        context.Context
		cancelFunc context.CancelFunc
		processes  []ManagedProcess
		wg         *sync.WaitGroup
		sigChan    chan os.Signal
		log        Logger
	}
)

// TakeOnControll takes the process under ExecutionManager controll.
// It's Run method will be executed when ExecuteProcesses is called.
func (r *ExecutionManager) TakeOnControll(process ...ManagedProcess) {
	r.processes = append(r.processes, process...)
}

// ExecuteProcesses starts all processes under its control by calling the Run method.
// If one of the conditions is met:
//  - Programm receives one of the predefined signals;
//  - One of the processes returns from its Run method,
// the ctx of all processes will be canceled, so every Run method should return so fast as it can.
// This method returns when all controlled processes returns from their Run methods.
func (r *ExecutionManager) ExecuteProcesses() {
	defer r.cancelFunc()
	r.logInfof("start all processes execution")
	defer r.logInfof("all processes execution stopped")
	r.wg.Add(len(r.processes))
	for _, process := range r.processes {
		go r.run(process)
	}
	go r.listenSignals()
	r.wg.Wait()
}

func (r *ExecutionManager) listenSignals() {
	select {
	case signal := <-r.sigChan:
		r.logInfof("graceful shutdown started by signal %q", signal.String())
		r.cancelFunc()
	case <-r.ctx.Done():
		break
	}
}

func (r *ExecutionManager) run(process ManagedProcess) {
	defer r.wg.Done()
	r.logInfof("process %q started", process.Name())
	defer r.logInfof("process %q stopped", process.Name())
	defer func() {
		if val := recover(); val != nil {
			r.logErrorf("panic detected in process %q: %v", process.Name(), val)
		}
		if r.ctx.Err() == nil {
			r.logInfof("graceful shutdown started by process %q", process.Name())
			r.cancelFunc()
		}
	}()
	process.Run(r.ctx)
}

func (r *ExecutionManager) logInfof(format string, args ...interface{}) {
	if r.log != nil {
		r.log.Infof(format, args...)
	}
}

func (r *ExecutionManager) logErrorf(format string, args ...interface{}) {
	if r.log != nil {
		r.log.Errorf(format, args...)
	}
}
