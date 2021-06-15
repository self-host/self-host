/*
Copyright © 2021 Self-host Authors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/self-host/self-host/pkg/util/templates"
)

var (
	programCmdLong = templates.Examples(`
		Develop Self-host programs locally
	`)

	programCmdExample = templates.Examples(`
		# Compile a program to check for errors.
		selfctl program compile -l tengo -f myprog.tengo

		# Run a program
		selfctl program run -l tengo -f myprog.tengo
	`)
)

var programCmd = &cobra.Command{
	Use:     "program",
	Short:   "Develop Self-host programs locally",
	Long:    programCmdLong,
	Example: programCmdExample,
}

var (
	programLanguage string
	programFilename string
	programDeadline time.Duration
)

func init() {
	rootCmd.AddCommand(programCmd)
}
