// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package templates

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

const indentation = `  `

type normalizer struct {
	string
}

// LongDesc normalizes a command's long description to follow the conventions.
func LongDesc(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.heredoc().trim().string
}

// Examples normalizes a command's examples to follow the conventions.
func Examples(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.trim().indent().string
}

// Normalize perform all required normalizations on a given command.
func Normalize(cmd *cobra.Command) *cobra.Command {
	if len(cmd.Example) > 0 {
		cmd.Example = Examples(cmd.Example)
	}
	if len(cmd.Long) > 0 {
		cmd.Long = LongDesc(cmd.Long)
	}
	return cmd
}

func (n normalizer) heredoc() normalizer {
	n.string = heredoc.Doc(n.string)
	return n
}

func (n normalizer) trim() normalizer {
	n.string = strings.TrimSpace(n.string)
	return n
}

func (n normalizer) indent() normalizer {
	indentedLines := []string{}
	for _, line := range strings.Split(n.string, "\n") {
		trimmed := strings.TrimSpace(line)
		indented := indentation + trimmed
		indentedLines = append(indentedLines, indented)
	}
	n.string = strings.Join(indentedLines, "\n")
	return n
}
