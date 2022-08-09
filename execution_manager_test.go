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

package execution_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/goofinator/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxProcessDuration time.Duration = 5 * time.Second

func newSomeProcess(manualStopCh <-chan struct{}, wg *sync.WaitGroup) *someProcess {
	wg.Add(1)
	return &someProcess{
		manualStopCh: manualStopCh,
		runCalledCh:  make(chan struct{}),
		wg:           wg,
	}
}

type someProcess struct {
	wg              *sync.WaitGroup
	manualStopCh    <-chan struct{}
	runCalledCh     chan struct{}
	isExitByCommand bool
	isExitByContext bool
}

func (r someProcess) isRunCalledCh() <-chan struct{} { return r.runCalledCh }

func (r *someProcess) Run(ctx context.Context) {
	close(r.runCalledCh)
	defer r.wg.Done()
	select {
	case <-r.manualStopCh:
		r.isExitByCommand = true
		break
	case <-ctx.Done():
		r.isExitByContext = true
		break
	case <-time.After(maxProcessDuration):
		break
	}
}

func TestExecutionManagerWithNoProcess(t *testing.T) {
	wg := sync.WaitGroup{}
	reportCh := make(chan struct{})

	executionManager := execution.NewDefaultExecutionManager()
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	wg.Wait()
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
}

func TestExecutionManagerWithOneProcessByItsExit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess(stopCh, &wg)

	executionManager := execution.NewDefaultExecutionManager()
	executionManager.TakeOnControll(process1)
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	require.NoError(t, whaitChan(process1.isRunCalledCh()),
		"expect: every process.Run method is called\ngot: not every process.Run method is called")
	assert.True(t, isChanBlocked(reportCh),
		"expect: ExecuteProcesses waits for all processes shutdown\ngot: ExecuteProcesses returns before any process.Run has been returned")
	close(stopCh)
	wg.Wait()

	assert.True(t, process1.isExitByCommand)
	assert.False(t, process1.isExitByContext)
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
}

func TestExecutionManagerWithTwoProcessesByProcess1Exit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess(stopCh, &wg)
	process2 := newSomeProcess(nil, &wg)

	executionManager := execution.NewDefaultExecutionManager()
	executionManager.TakeOnControll(process1, process2)
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	require.NoError(t, whait2Chans(process1.isRunCalledCh(), process2.isRunCalledCh()),
		"expect: every process.Run method is called\ngot: not every process.Run method is called")
	assert.True(t, isChanBlocked(reportCh),
		"expect: ExecuteProcesses waits for all processes shutdown\ngot: ExecuteProcesses returns before any process.Run has been returned")
	close(stopCh)
	wg.Wait()

	assert.True(t, process1.isExitByCommand)
	assert.False(t, process2.isExitByCommand)
	assert.False(t, process1.isExitByContext)
	assert.True(t, process2.isExitByContext)
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
}

func TestExecutionManagerWithTwoProcessesByProcess2Exit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess(nil, &wg)
	process2 := newSomeProcess(stopCh, &wg)

	executionManager := execution.NewDefaultExecutionManager()
	executionManager.TakeOnControll(process1)
	executionManager.TakeOnControll(process2)
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	require.NoError(t, whait2Chans(process1.isRunCalledCh(), process2.isRunCalledCh()),
		"expect: every process.Run method is called\ngot: not every process.Run method is called")
	assert.True(t, isChanBlocked(reportCh),
		"expect: ExecuteProcesses waits for all processes shutdown\ngot: ExecuteProcesses returns before any process.Run has been returned")
	close(stopCh)
	wg.Wait()

	assert.False(t, process1.isExitByCommand)
	assert.True(t, process2.isExitByCommand)
	assert.True(t, process1.isExitByContext)
	assert.False(t, process2.isExitByContext)
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
}

func TestExecutionManagerWithTwoProcessesBySignalExit(t *testing.T) {
	wg := sync.WaitGroup{}
	reportCh := make(chan struct{})
	stopSignal := syscall.SIGUSR1
	process1 := newSomeProcess(nil, &wg)
	process2 := newSomeProcess(nil, &wg)

	config := execution.ExecutionManagerConfig{
		Signals: []os.Signal{stopSignal},
	}
	executionManager := execution.NewExecutionManager(config)
	executionManager.TakeOnControll(process2, process1)
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	require.NoError(t, whait2Chans(process1.isRunCalledCh(), process2.isRunCalledCh()),
		"expect: every process.Run method is called\ngot: not every process.Run method is called")
	assert.True(t, isChanBlocked(reportCh),
		"expect: ExecuteProcesses waits for all processes shutdown\ngot: ExecuteProcesses returns before any process.Run has been returned")
	require.NoError(t, syscall.Kill(syscall.Getpid(), stopSignal))
	wg.Wait()

	assert.False(t, process1.isExitByCommand)
	assert.False(t, process2.isExitByCommand)
	assert.True(t, process1.isExitByContext)
	assert.True(t, process2.isExitByContext)
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
}

func isChanBlocked(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return false
	default:
		return true
	}
}

func whaitChan(ch <-chan struct{}) error {
	select {
	case <-ch:
		return nil
	case <-time.After(maxProcessDuration):
		return fmt.Errorf("whaiting on chanel failed by timeout")
	}
}

func whait2Chans(chan1 <-chan struct{}, chan2 <-chan struct{}) error {
	for i := 0; i < 2; i++ {
		select {
		case <-chan1:
			chan1 = nil
		case <-chan2:
			chan2 = nil
		case <-time.After(maxProcessDuration):
			return fmt.Errorf("whaiting on 2 chanels failed by timeout: chan1 is Done: %t, chan2 is Done: %t", chan1 == nil, chan2 == nil)
		}
	}
	return nil
}
