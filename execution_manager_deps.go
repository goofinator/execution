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

import "context"

type (
	// ManagedProcess implements some process within it's Run method.
	// Run method should return as the ctx is cancelled.
	ManagedProcess interface {
		Run(ctx context.Context)
		Name() string
	}
	// Logger to write processes execution status log.
	// It should usable with concurrent writing.
	Logger interface {
		Infof(format string, args ...interface{})
		Errorf(format string, args ...interface{})
	}
)
