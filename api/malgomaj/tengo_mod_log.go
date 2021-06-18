// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

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
