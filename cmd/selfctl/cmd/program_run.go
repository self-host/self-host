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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/self-host/self-host/api/malgomaj"
	"github.com/self-host/self-host/pkg/util/templates"
	"github.com/spf13/cobra"
)

var (
	runCmdLong = templates.Examples(`
		Run a program
	`)

	runCmdExample = templates.Examples(`
		# Run a program
		selfctl program run -l tengo -f myprog.tengo
	`)
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run -f FILENAME",
	Short:   "Run a program",
	Long:    runCmdLong,
	Example: runCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		if programLanguage != "tengo" {
			fmt.Fprintln(os.Stderr, "unsupported language:", programLanguage)
			os.Exit(1)
		}

		source_code, err := ioutil.ReadFile(programFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		program, err := malgomaj.NewProgram("selfctl", programFilename, programLanguage, programDeadline, source_code)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = program.Run(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	programCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&programFilename, "filename", "f", "", "Program source file")
	runCmd.Flags().StringVarP(&programLanguage, "lang", "l", "tengo", "Program language")
	runCmd.Flags().DurationVarP(&programDeadline, "deadline", "", 5*time.Second, "Program deadline")
	runCmd.MarkFlagRequired("filename")
}
