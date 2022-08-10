# execution

This package provides a kind of framework to run a few of hi level goroutine which are locally called processes
and provide it's graceful shutdown by context cancellation.

If your process object has a methods to satisfy **execution.ManagedProcess** interface.
It could be taken on the control with **execution.ExecutionManager** by calling it's **TakeOnControll** method.

All of these processes will be executed in a parallel goroutines by the **ExecuteProcesses** method of **execution.ExecutionManager** call.
If on of these goroutines will return, or if an application will get a configured signal from OS 
(SIGINT or SIGTERM by default) the common context will be cancelled.
**ExecuteProcesses** method will return when all of processes under control returns from it's **Run** method.

You can provide a configuration of type **execution.ExecutionManagerConfig** to setup a set of signals to start a procedure of graceful shutdown.
Also external logger could be provided to display a log of execution state changing. This logger should satisfy an **execution.Logger** interface.

Example of using this framework to manage the execution of two web servers available [here](examples/few_web_servers/README.md).