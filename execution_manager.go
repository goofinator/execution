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
	"sync"
)

func NewExecutionManager() *ExecutionManager {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &ExecutionManager{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		wg:         &sync.WaitGroup{},
	}
}

// The ExecutionManager ensures that multiple processes are executed in a controlled manner.
// A graceful termination will be initiated if one of the conditions is met:
//  - ExecutionManager receives a SIGINT or SIGTERM signal.
//  - One of the processes returns from its Run method
type ExecutionManager struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	processes  []ManagedProcess
	wg         *sync.WaitGroup
}

// TakeOnControll takes process under ExecutionManager controll.
// It's Run method will be executed when ExecuteProcesses is called.
func (r *ExecutionManager) TakeOnControll(process ...ManagedProcess) {
	r.processes = append(r.processes, process...)
}

// ExecuteProcesses starts all processes under its control by calling the Run method.
// If one of the conditions is met:
//  - ExecutionManager receives a SIGINT or SIGTERM signal;
//  - One of the processes returns from its Run method;
// the ctx of all processes will be canceled, so every Run method should return so fast as it can.
// This method returns when all controlled processes returns from their Run methods.
func (r ExecutionManager) ExecuteProcesses() {
	defer r.cancelFunc()
	r.wg.Add(len(r.processes))
	for _, process := range r.processes {
		go r.run(process)
	}
	r.wg.Wait()
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
