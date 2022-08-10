<!--
 Copyright (C) 2022 Blinnikov AA <goofinator@mail.ru>
 
 This file is part of execution.
 
 execution is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.
 
 execution is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.
 
 You should have received a copy of the GNU General Public License
 along with execution.  If not, see <http://www.gnu.org/licenses/>.
-->

# example of execution.ExecutionManager using

If you execute this example with 
```bash
go run examples/few_web_servers/cmd/app.go
```
and send a few **Get** requests from another terminal
```bash
curl -G http://localhost:8080/api/v1/wisdom
curl -G http://localhost:5000/health-check
```
and terminate the application with SIGTERM (Ctrl+C in the first terminal),
you'll see the log of two web services execution managment.
```bash
INFO[0000] start all processes execution                
INFO[0000] process "api web server" started             
INFO[0000] process "monitoring web server" started      
INFO[0002] GET: /api/v1/wisdom                          
INFO[0004] GET: /api/v1/ping                            
^CINFO[0007] graceful shutdown started by signal "interrupt" 
INFO[0007] process "monitoring web server" stopped      
INFO[0007] process "api web server" stopped             
INFO[0007] all processes execution stopped
```
