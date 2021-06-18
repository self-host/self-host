// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package rest

type NewPolicyAction PolicyAction
type NewPolicyEffect PolicyEffect

type UpdatePolicyAction PolicyAction
type UpdatePolicyEffect PolicyEffect

type NewDatasetFormat DatasetFormat
type UpdateDatasetFormat DatasetFormat

type UpdateThingState ThingState

type NewProgramLanguage ProgramLanguage
type NewProgramState ProgramState
type NewProgramType ProgramType

type UpdateProgramLanguage ProgramLanguage
type UpdateProgramState ProgramState
type UpdateProgramType ProgramType

func GetOpenAPIFile() ([]byte, error) {
	return decodeSpec()
}
