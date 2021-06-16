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