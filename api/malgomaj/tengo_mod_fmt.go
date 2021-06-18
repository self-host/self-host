// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"github.com/d5/tengo/v2"
)

var fmtModule = map[string]tengo.Object{
	"sprintf": &tengo.UserFunction{Name: "sprintf", Value: fmtSprintf},
}

func fmtSprintf(args ...tengo.Object) (ret tengo.Object, err error) {
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
		// tengo.String is immutable
		return format, nil
	}
	s, err := tengo.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}

	return &tengo.String{Value: s}, nil
}
