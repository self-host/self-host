// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/d5/tengo/v2"
)

type Response struct {
	StatusCode int
	Body       []byte
}

const (
	tengoHTTPClient = "Selfhost-tengo-client/1.0"
)

var httpModule = map[string]tengo.Object{
	"get":        &tengo.UserFunction{Name: "get", Value: httpGet},
	"put":        &tengo.UserFunction{Name: "put", Value: httpPut},
	"post":       &tengo.UserFunction{Name: "post", Value: httpPost},
	"postform":   &tengo.UserFunction{Name: "postform", Value: httpPostForm},
	"delete":     &tengo.UserFunction{Name: "delete", Value: httpDelete},
	"toformdata": &tengo.UserFunction{Name: "toformdata", Value: httpToFormData},
}

func httpRespToTengo(resp *http.Response) (tengo.Object, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	robj := make(map[string]tengo.Object)
	robj["StatusCode"] = &tengo.Int{Value: int64(resp.StatusCode)}
	robj["Body"] = &tengo.Bytes{Value: body}
	robj["ContentLength"] = &tengo.Int{Value: int64(len(body))}
	robj["Header"] = &tengo.Map{Value: make(map[string]tengo.Object)}

	for key, vals := range resp.Header {
		hdr, ok := robj["Header"].(*tengo.Map)
		if ok == false {
			// ERROR OUT
			continue
		}

		tVals := tengo.Array{}
		for _, val := range vals {
			tVals.Value = append(tVals.Value, &tengo.String{Value: val})
		}

		hdr.Value[key] = &tVals
	}

	return tengo.FromInterface(robj)
}

func httpGet(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, tengo.ErrWrongNumArguments
	}

	url, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "url",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	req.Header.Set("User-Agent", tengoHTTPClient)

	if len(args) >= 2 && args[1] != tengo.UndefinedValue {
		tobj := tengo.ToInterface(args[1])
		queryArgs, ok := tobj.([]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "query_args",
				Expected: "[]string",
				Found:    args[1].TypeName(),
			}
		}

		q := req.URL.Query()
		for index, val := range queryArgs {
			sval, ok := val.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    "unknown",
				}
			}

			vals := strings.SplitN(sval, "=", 2)
			if len(vals) != 2 {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    sval,
				}
			}

			q.Add(vals[0], vals[1])
		}

		req.URL.RawQuery = q.Encode()
	}

	if len(args) >= 3 {
		tobj := tengo.ToInterface(args[2])
		headers, ok := tobj.(map[string]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "headers",
				Expected: "map[string]string",
				Found:    args[2].TypeName(),
			}
		}

		for key, element := range headers {
			val, ok := element.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("headers[%s]", key),
					Expected: "string",
					Found:    "unknown",
				}
			}
			req.Header.Set(key, val)
		}
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}
	defer resp.Body.Close()

	tresp, err := httpRespToTengo(resp)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	return tresp, nil
}

func httpDoWithBody(method string, args ...tengo.Object) (ret tengo.Object, err error) {
	// Format is "determined" by Header: Content-Type

	if len(args) < 1 || len(args) > 4 {
		return nil, tengo.ErrWrongNumArguments
	}

	url, ok := tengo.ToString(args[0])
	if ok == false {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "url",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	var submitBody string
	if len(args) == 4 && args[3] != tengo.UndefinedValue {
		submitBody, ok = tengo.ToString(args[3])
		if ok == false {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "body",
				Expected: "string",
				Found:    args[3].TypeName(),
			}
		}
	}

	req, err := http.NewRequest(method, url, strings.NewReader(submitBody))
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	req.Header.Set("User-Agent", tengoHTTPClient)

	if len(args) >= 2 && args[1] != tengo.UndefinedValue {
		tobj := tengo.ToInterface(args[1])
		queryArgs, ok := tobj.([]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "query_args",
				Expected: "[]string",
				Found:    args[1].TypeName(),
			}
		}

		q := req.URL.Query()
		for index, val := range queryArgs {
			sval, ok := val.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    "unknown",
				}
			}

			vals := strings.SplitN(sval, "=", 2)
			if len(vals) != 2 {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    sval,
				}
			}

			q.Add(vals[0], vals[1])
		}

		req.URL.RawQuery = q.Encode()
	}

	if len(args) >= 3 {
		tobj := tengo.ToInterface(args[2])
		headers, ok := tobj.(map[string]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "headers",
				Expected: "map[string]string",
				Found:    args[2].TypeName(),
			}
		}

		for key, element := range headers {
			val, ok := element.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("headers[%s]", key),
					Expected: "string",
					Found:    "unknown",
				}
			}
			req.Header.Set(key, val)
		}
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}
	defer resp.Body.Close()

	tresp, err := httpRespToTengo(resp)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	return tresp, nil
}

func httpDelete(args ...tengo.Object) (ret tengo.Object, err error) {
	return httpDoWithBody("DELETE", args...)
}

func httpPost(args ...tengo.Object) (ret tengo.Object, err error) {
	return httpDoWithBody("POST", args...)
}

func httpPut(args ...tengo.Object) (ret tengo.Object, err error) {
	return httpDoWithBody("PUT", args...)
}

func httpToFormData(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) < 1 || len(args) > 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	tobj := tengo.ToInterface(args[0])
	queryArgs, ok := tobj.([]interface{})
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "query_args",
			Expected: "[]string",
			Found:    args[0].TypeName(),
		}
	}

	values := url.Values{}
	for index, val := range queryArgs {
		sval, ok := val.(string)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("query_args[%v]", index),
				Expected: "key=value",
				Found:    "unknown",
			}
		}

		vals := strings.SplitN(sval, "=", 2)
		if len(vals) != 2 {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("query_args[%v]", index),
				Expected: "key=value",
				Found:    sval,
			}
		}
		values.Set(vals[0], vals[1])
	}

	return tengo.FromInterface(values.Encode())
}

func httpPostForm(args ...tengo.Object) (ret tengo.Object, err error) {
	// Query string will be encoded as form data
	//
	// application/x-www-form-urlencoded

	values := url.Values{}

	if len(args) < 1 || len(args) > 3 {
		return nil, tengo.ErrWrongNumArguments
	}

	url, ok := tengo.ToString(args[0])
	if ok == false {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "url",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	if len(args) >= 2 && args[1] != tengo.UndefinedValue {
		tobj := tengo.ToInterface(args[1])
		queryArgs, ok := tobj.([]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "query_args",
				Expected: "[]string",
				Found:    args[1].TypeName(),
			}
		}

		for index, val := range queryArgs {
			sval, ok := val.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    "unknown",
				}
			}

			vals := strings.SplitN(sval, "=", 2)
			if len(vals) != 2 {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("query_args[%v]", index),
					Expected: "key=value",
					Found:    sval,
				}
			}

			values.Set(vals[0], vals[1])
		}
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(values.Encode()))
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	req.Header.Set("User-Agent", tengoHTTPClient)

	if len(args) >= 3 {
		tobj := tengo.ToInterface(args[2])
		headers, ok := tobj.(map[string]interface{})
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "headers",
				Expected: "map[string]string",
				Found:    args[2].TypeName(),
			}
		}

		for key, element := range headers {
			val, ok := element.(string)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("headers[%s]", key),
					Expected: "string",
					Found:    "unknown",
				}
			}
			req.Header.Set(key, val)
		}
	}

	// Must be application/x-www-form-urlencoded
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}
	defer resp.Body.Close()

	tresp, err := httpRespToTengo(resp)
	if err != nil {
		// Return error to Tengo and do not cause Runtime Error
		return &tengo.Error{
			Value: &tengo.String{Value: err.Error()},
		}, nil
	}

	return tresp, nil
}
