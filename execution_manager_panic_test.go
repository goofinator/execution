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
	"sync"
	"testing"
	"time"

	"github.com/goofinator/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPanicProcess(name string, manualStopCh <-chan struct{}, wg *sync.WaitGroup) *panicProcess {
	wg.Add(1)
	return &panicProcess{
		manualStopCh: manualStopCh,
		runCalledCh:  make(chan struct{}),
		wg:           wg,
		name:         name,
	}
}

type panicProcess struct {
	name            string
	wg              *sync.WaitGroup
	manualStopCh    <-chan struct{}
	runCalledCh     chan struct{}
	isExitByCommand bool
	isExitByContext bool
}

func (r *panicProcess) isRunCalledCh() <-chan struct{} { return r.runCalledCh }

func (r panicProcess) Name() string { return r.name }

func (r *panicProcess) Run(ctx context.Context) {
	close(r.runCalledCh)
	defer r.wg.Done()
	select {
	case <-r.manualStopCh:
		r.isExitByCommand = true
		panic("AAA!!! Aliens!!! Not again!!!")
	case <-ctx.Done():
		r.isExitByContext = true
	case <-time.After(maxProcessDuration):
		break
	}
}

func TestExecutionManagerWithOnePanicProcess(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newPanicProcess("process1", stopCh, &wg)

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

func TestExecutionManagerWithOnePanicProcessAndLogger(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newPanicProcess("process1", stopCh, &wg)

	log := newLogger()
	config := execution.ExecutionManagerConfig{
		Log: log,
	}
	executionManager := execution.NewExecutionManager(config)
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

	require.Equal(t, 6, len(log.lines))
	require.Equal(t, "INFO: start all processes execution", log.lines[0])
	require.Equal(t, "INFO: process \"process1\" started", log.lines[1])
	require.Equal(t, "ERROR: panic detected in process \"process1\": AAA!!! Aliens!!! Not again!!!", log.lines[2])
	require.Equal(t, "INFO: graceful shutdown started by process \"process1\"", log.lines[3])
	require.Equal(t, "INFO: process \"process1\" stopped", log.lines[4])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[5])
}

func TestExecutionManagerWithTwoProcessesByPanic(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newPanicProcess("process1", stopCh, &wg)
	process2 := newSomeProcess("process2", nil, &wg)

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

func TestExecutionManagerWithTwoProcessesAndLoggerByPanic(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newPanicProcess("process1", stopCh, &wg)
	process2 := newSomeProcess("process2", nil, &wg)

	log := newLogger()
	config := execution.ExecutionManagerConfig{
		Log: log,
	}
	executionManager := execution.NewExecutionManager(config)
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

	require.Equal(t, 8, len(log.lines))
	assert.Equal(t, "INFO: start all processes execution", log.lines[0])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" started", "INFO: process \"process2\" started"},
		log.lines[1:3])
	assert.Equal(t, "ERROR: panic detected in process \"process1\": AAA!!! Aliens!!! Not again!!!", log.lines[3])
	assert.Equal(t, "INFO: graceful shutdown started by process \"process1\"", log.lines[4])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" stopped", "INFO: process \"process2\" stopped"},
		log.lines[5:7])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[7])
}
