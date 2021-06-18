// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"bytes"
	"github.com/d5/tengo/v2"
)

type cgiModule struct {
	tengo.ObjectImpl
	// http.Request
	reqHeaders map[string]string
	reqBody    []byte

	respBody    bytes.Buffer
	respHeaders map[string]string
	respStatus  int
}

func (cgi *cgiModule) Import(moduleName string) (interface{}, error) {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"request": &tengo.ImmutableMap{
				Value: map[string]tengo.Object{
					"headers": &tengo.UserFunction{
						Name:  "headers",
						Value: cgi.GetHeaders,
					},
					"body": &tengo.UserFunction{
						Name:  "body",
						Value: cgi.GetBody,
					},
				},
			},
			"response": &tengo.ImmutableMap{
				Value: map[string]tengo.Object{
					"headers": &tengo.UserFunction{
						Name:  "headers",
						Value: cgi.SetHeader,
					},
					"write": &tengo.UserFunction{
						Name:  "write",
						Value: cgi.WriteBody,
					},
					"status": &tengo.UserFunction{
						Name:  "status",
						Value: cgi.SetStatus,
					},
				},
			},
		},
	}, nil
}

func (cgi *cgiModule) GetHeaders(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	return tengo.FromInterface(cgi.reqHeaders)
}

func (cgi *cgiModule) SetHeader(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	v, ok := args[0].(*tengo.String)
	if ok == false {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "key",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	header := v.Value

	v, ok = args[1].(*tengo.String)
	if ok == false {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	cgi.respHeaders[header] = v.Value

	return nil, nil
}

func (cgi *cgiModule) SetStatus(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	v, ok := args[0].(*tengo.Int)
	if ok == false {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "status",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}
	cgi.respStatus = int(v.Value)

	return nil, nil
}

func (cgi *cgiModule) GetBody(args ...tengo.Object) (ret tengo.Object, err error) {
	return &tengo.Bytes{
		Value: cgi.reqBody,
	}, nil
}

func (cgi *cgiModule) WriteBody(args ...tengo.Object) (ret tengo.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	if numArgs > 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	var b []byte

	switch v := args[0].(type) {
	case *tengo.String:
		b = []byte(v.Value)
	case *tengo.Bytes:
		b = v.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "data",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	count, err := cgi.respBody.Write(b)
	if err != nil {
		return nil, err
	}

	return &tengo.Int{Value: int64(count)}, nil
}
