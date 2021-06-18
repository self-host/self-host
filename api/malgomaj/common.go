// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"fmt"
)

type NewTask_Http_Headers map[string]string
type NewTask_Libraries map[string][]byte

type NewTaskLanguage string

// Defines values for NewTaskLanguage.
const (
	NewTaskLanguageTengo NewTaskLanguage = "tengo"
)

type NewTaskHttp struct {
	Body    []byte               `json:"body"`
	Headers NewTask_Http_Headers `json:"headers"`
}

// Return the Id used by the Cache
func (t *NewTask) GetId() string {
	return fmt.Sprintf("%v/%v", t.Domain, t.ProgramUuid)
}

func GetOpenAPIFile() ([]byte, error) {
	return decodeSpec()
}
