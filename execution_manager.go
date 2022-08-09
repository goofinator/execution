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
	// Signals - is a list of syscall.Signals capable to initiate shotdown (default: [SIGINT, SIGTERM])
	Signals []os.Signal
}

// NewDefaultExecutionManager creates preconfigured ExecutionManager
// gracefull shutdown will be initiated by SIGINT or SIGTERM signals
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
	}
}

// The ExecutionManager ensures that multiple processes are executed in a controlled manner.
// A graceful termination will be initiated if one of the conditions is met:
//  - Programm receives one of the predefined signals.
//  - One of the processes returns from its Run method
type ExecutionManager struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	processes  []ManagedProcess
	wg         *sync.WaitGroup
	sigChan    chan os.Signal
}

// TakeOnControll takes process under ExecutionManager controll.
// It's Run method will be executed when ExecuteProcesses is called.
func (r *ExecutionManager) TakeOnControll(process ...ManagedProcess) {
	r.processes = append(r.processes, process...)
}

// ExecuteProcesses starts all processes under its control by calling the Run method.
// If one of the conditions is met:
//  - Programm receives one of the predefined signals;
//  - One of the processes returns from its Run method;
// the ctx of all processes will be canceled, so every Run method should return so fast as it can.
// This method returns when all controlled processes returns from their Run methods.
func (r ExecutionManager) ExecuteProcesses() {
	defer r.cancelFunc()
	if len(r.processes) == 0 {
		return
	}
	r.wg.Add(len(r.processes) + 1)
	go r.listenSignals()
	for _, process := range r.processes {
		go r.run(process)
	}
	r.wg.Wait()
}

func (r ExecutionManager) listenSignals() {
	defer r.wg.Done()
	select {
	case <-r.sigChan:
		r.cancelFunc()
	case <-r.ctx.Done():
		break
	}
}

func (r ExecutionManager) run(process ManagedProcess) {
	defer r.wg.Done()
	process.Run(r.ctx)
	if r.ctx.Err() == nil {
		r.cancelFunc()
	}
}

// ManagedProcess implements some process within it's Run method.
// Run method should return as the ctx is cancelled.
type ManagedProcess interface {
	Run(ctx context.Context)
}
