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

package templates

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

const Indentation = `  `

type normalizer struct {
	string
}

func LongDesc(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.heredoc().trim().string
}

func Examples(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.trim().indent().string
}

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
		indented := Indentation + trimmed
		indentedLines = append(indentedLines, indented)
	}
	n.string = strings.Join(indentedLines, "\n")
	return n
}
