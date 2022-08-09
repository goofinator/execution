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
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/goofinator/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionManagerWithNoProcessAndLogger(t *testing.T) {
	wg := sync.WaitGroup{}
	reportCh := make(chan struct{})

	log := newLogger()
	config := execution.ExecutionManagerConfig{
		Log: log,
	}
	executionManager := execution.NewExecutionManager(config)
	go func() {
		executionManager.ExecuteProcesses()
		close(reportCh)
	}()

	wg.Wait()
	require.NoError(t, whaitChan(reportCh), "expect: ExecuteProcesses exit\ngot: ExecuteProcesses hanging")
	require.Equal(t, 2, len(log.lines))
	require.Equal(t, "INFO: start all processes execution", log.lines[0])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[1])
}

func TestExecutionManagerWithOneProcessAndLoggerByItsExit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess("process1", stopCh, &wg)

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

	require.Equal(t, 5, len(log.lines))
	assert.Equal(t, "INFO: start all processes execution", log.lines[0])
	assert.Equal(t, "INFO: process \"process1\" started", log.lines[1])
	assert.Equal(t, "INFO: graceful shutdown started by process \"process1\"", log.lines[2])
	assert.Equal(t, "INFO: process \"process1\" stopped", log.lines[3])
	assert.Equal(t, "INFO: all processes execution stopped", log.lines[4])
}

func TestExecutionManagerWithTwoProcessesAndLoggerByProcess1Exit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess("process1", stopCh, &wg)
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

	require.Equal(t, 7, len(log.lines))
	assert.Equal(t, "INFO: start all processes execution", log.lines[0])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" started", "INFO: process \"process2\" started"},
		log.lines[1:3])
	assert.Equal(t, "INFO: graceful shutdown started by process \"process1\"", log.lines[3])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" stopped", "INFO: process \"process2\" stopped"},
		log.lines[4:6])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[6])
}

func TestExecutionManagerWithTwoProcessesAndLoggerByProcess2Exit(t *testing.T) {
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	reportCh := make(chan struct{})
	process1 := newSomeProcess("process1", nil, &wg)
	process2 := newSomeProcess("process2", stopCh, &wg)

	log := newLogger()
	config := execution.ExecutionManagerConfig{
		Log: log,
	}
	executionManager := execution.NewExecutionManager(config)
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

	require.Equal(t, 7, len(log.lines))
	assert.Equal(t, "INFO: start all processes execution", log.lines[0])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" started", "INFO: process \"process2\" started"},
		log.lines[1:3])
	assert.Equal(t, "INFO: graceful shutdown started by process \"process2\"", log.lines[3])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" stopped", "INFO: process \"process2\" stopped"},
		log.lines[4:6])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[6])
}

func TestExecutionManagerWithTwoProcessesAndLoggerBySignalExit(t *testing.T) {
	wg := sync.WaitGroup{}
	reportCh := make(chan struct{})
	stopSignal := syscall.SIGUSR2
	process1 := newSomeProcess("process1", nil, &wg)
	process2 := newSomeProcess("process2", nil, &wg)

	log := newLogger()
	config := execution.ExecutionManagerConfig{
		Signals: []os.Signal{stopSignal},
		Log:     log,
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

	require.Equal(t, 7, len(log.lines))
	assert.Equal(t, "INFO: start all processes execution", log.lines[0])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" started", "INFO: process \"process2\" started"},
		log.lines[1:3])
	assert.Equal(t, "INFO: graceful shutdown started by signal \"user defined signal 2\"", log.lines[3])
	assert.ElementsMatch(t,
		[]string{"INFO: process \"process1\" stopped", "INFO: process \"process2\" stopped"},
		log.lines[4:6])
	require.Equal(t, "INFO: all processes execution stopped", log.lines[6])
}

func newLogger() *logger { return &logger{lines: []string{}} }

type logger struct {
	lines  []string
	locker sync.Mutex
}

func (r *logger) Infof(format string, args ...interface{}) {
	r.locker.Lock()
	defer r.locker.Unlock()
	line := fmt.Sprintf("INFO: "+format, args...)
	r.lines = append(r.lines, line)
}

func (r *logger) Errorf(format string, args ...interface{}) {
	r.locker.Lock()
	defer r.locker.Unlock()
	line := fmt.Sprintf("ERROR: "+format, args...)
	r.lines = append(r.lines, line)
}
