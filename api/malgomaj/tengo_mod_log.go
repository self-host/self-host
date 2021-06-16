/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/
package malgomaj

import (
	"github.com/d5/tengo/v2"
	"go.uber.org/zap"
)

var logModuleLogger *zap.Logger

func init() {
	var err error
	logModuleLogger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}
}

var logModule = map[string]tengo.Object{
	"info":  &tengo.UserFunction{Name: "info", Value: logInfo},
	"error": &tengo.UserFunction{Name: "error", Value: logInfo},
}

func logInfo(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	format, ok := args[0].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	if numArgs == 1 {
		logModuleLogger.Info(format.Value)
		return nil, nil
	}

	s, err := tengo.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}

	logModuleLogger.Info(s)

	return nil, nil
}

func logError(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	format, ok := args[0].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	if numArgs == 1 {
		logModuleLogger.Error(format.Value)
		return nil, nil
	}

	s, err := tengo.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}

	logModuleLogger.Error(s)

	return nil, nil
}
