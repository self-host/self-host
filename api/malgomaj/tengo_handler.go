// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"

	"github.com/self-host/self-host/api/malgomaj/library"
	"github.com/self-host/self-host/pkg/util"
)

var AllowedBaseModules = []string{
	"base64",
	"enum", // source module
	// "fmt", disabled
	"hex",
	"json",
	"math",
	"rand",
	"text",
	"times",
}

var ExtraModules = map[string]map[string]tengo.Object{
	"fmt":  fmtModule, // Our own fmt module without Print* functions
	"http": httpModule,
	"log":  logModule,
	"cgi":  nil, // Initialize on every call
}

var tengoImportRegex = regexp.MustCompile(`import\("(.*)"\)`)

func moduleMapKeys(m map[string]map[string]tengo.Object) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

type TengoProgram struct {
	sync.RWMutex

	domain     string
	id         string
	deadline   time.Duration
	sourceCode []byte
	byteCode   *tengo.Compiled // Compiled code

	cgi *cgiModule
}

func NewTengoProgram(domain string, id string, deadline time.Duration, sourceCode []byte) *TengoProgram {
	return &TengoProgram{
		domain:     domain,
		id:         id,
		deadline:   deadline,
		sourceCode: sourceCode,
	}
}

func (p *TengoProgram) Language() string {
	return "tengo"
}

func (p *TengoProgram) Id() string {
	return fmt.Sprintf("%v/%v", p.domain, p.id)
}

func (p *TengoProgram) Deadline() time.Duration {
	return p.deadline
}

func (p *TengoProgram) Checksum() [16]byte {
	return md5.Sum(p.sourceCode)
}

func (p *TengoProgram) Equals(b Program) bool {
	return p.Id() == b.Id() &&
		p.Language() == b.Language() &&
		p.Deadline() == b.Deadline() &&
		p.Checksum() == b.Checksum()
}

func (p *TengoProgram) IsCGI() bool {
	return p.cgi != nil
}

func (p *TengoProgram) AllImports() []string {
	// Use a map to avoid duplicates
	m := make(map[string]struct{})

	for _, match := range tengoImportRegex.FindAllStringSubmatch(string(p.sourceCode), -1) {
		if len(match) == 2 {
			k := match[1]
			m[k] = struct{}{}
		}
	}

	imports := make([]string, 0)
	for k := range m {
		mod := k
		imports = append(imports, mod)
	}
	return imports
}

func (p *TengoProgram) Modules() []string {
	exkeys := moduleMapKeys(ExtraModules)
	libs := make([]string, 0)

	for _, imp := range p.AllImports() {
		if util.StringSliceContains(AllowedBaseModules, imp) == false &&
			util.StringSliceContains(exkeys, imp) == false {
			libs = append(libs, imp)
		}
	}

	return libs
}

func (p *TengoProgram) Compile(ctx context.Context) (err error) {
	modules := stdlib.GetModuleMap(AllowedBaseModules...)

	for name, mod := range ExtraModules {
		if name == "cgi" {
			p.cgi = &cgiModule{
				respHeaders: make(map[string]string),
			}
			modules.Add(name, p.cgi)
		} else {
			modules.AddBuiltinModule(name, mod)
		}
	}

	// All external modules declared in program source code
	for _, modname := range p.Modules() {
		name := modname
		revision := "latest"

		// Extract revision from modname
		sp := strings.Split(modname, "@")

		if len(sp) == 2 {
			name = sp[0]
			revision = sp[1]
		}

		// Get module from module library
		mod, err := library.Get(&library.LibraryParams{
			Domain:   p.domain,
			Module:   name,
			Revision: revision,
			Language: "tengo",
		})
		if err != nil {
			return err
		}

		// The module must maintain the name from the source code
		modules.AddSourceModule(modname, mod.Code)
	}

	script := tengo.NewScript(p.sourceCode)
	script.SetImports(modules)

	p.byteCode, err = script.Compile()
	return err
}

func (p *TengoProgram) Run(ctx context.Context) error {
	if p.byteCode == nil {
		if err := p.Compile(ctx); err != nil {
			return err
		}
	}

	lctx, cancel := context.WithTimeout(ctx, p.deadline)
	defer cancel() // Release context if execution finishes before deadline

	return p.byteCode.RunContext(lctx)
}

func (p *TengoProgram) RunWithHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if p.byteCode == nil {
		if err := p.Compile(ctx); err != nil {
			return err
		}
	}

	if p.IsCGI() == false {
		return errors.New("program can not handle a CGI request")
	}

	req, ok := ctx.Value("http").(*NewTaskHttp)
	if ok == false {
		return errors.New("http object was not provided to cgi program")
	}

	p.cgi.reqHeaders = req.Headers
	p.cgi.reqBody = req.Body

	lctx, cancel := context.WithTimeout(ctx, p.deadline)
	defer cancel() // Release context if execution finishes before deadline

	err := p.byteCode.RunContext(lctx)
	if err != nil {
		return err
	}

	w.WriteHeader(p.cgi.respStatus)

	for k, v := range p.cgi.respHeaders {
		w.Header().Set(k, v)
	}

	w.Write(p.cgi.respBody.Bytes())

	return nil
}
