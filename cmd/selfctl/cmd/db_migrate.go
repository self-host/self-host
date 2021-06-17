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
	"fmt"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/self-host/self-host/pkg/util/templates"
	"github.com/self-host/self-host/postgres"
	"github.com/spf13/cobra"
)

var (
	migrateCmdLong = templates.LongDesc(`
		handle Self-host database schema migration
	`)

	migrateCmdExample = templates.Examples(`
		# Initialize an empty database to the latest schema
		selfctl db migrate up --database URI

		# Upgrade a database to the latest schema
		selfctl db migrate up --database URI

		# Move 2 steps forward in the schema layout
		selfctrl db migrate up --steps 2 --database URI

		# Move 2 steps backward in the schema layout
		selfctrl db migrate down --steps 2 --database URI

		# Upgrade a database to a specific schema version
		selfctrl db migrate goto --version 12 --database URI

		# Downgrade a database to a specific schema version
		selfctrl db migrate goto --version 10 --database URI

		# Force set a certain schema version(dangerous)
		# Doesn't perform any up/down steps. Only sets the version.
		selfctrl db migrate force --version 11 --database URI
	`)
)

var (
	migrateCmd = &cobra.Command{
		Use:     "migrate [(up [N]|down [N]|goto N|force N)]",
		Short:   "handle Self-host database schema migration",
		Long:    migrateCmdLong,
		Example: migrateCmdExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mig, err := getMigrateInstance(migrateDbUri)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}

			switch args[0] {
			case "up":
				if migrateSteps == 0 {
					err = mig.Up()

					// FIXME: A (y/n) prompt
				} else {
					fmt.Println(int(migrateSteps))
					err = mig.Steps(int(migrateSteps))
				}

				if err != nil {
					if err.Error() == "file does not exist" {
						fmt.Println("Already on the latest version.")
					} else {
						fmt.Fprintln(os.Stderr, err.Error())
						os.Exit(1)
					}
				}

				dumpMigrateVersion(mig)

			case "down":
				if migrateSteps == 0 {
					err = mig.Down()

					// FIXME: A (y/n) prompt

					if err == nil {
						fmt.Println("Downgraded database to before the first version.")
					}
				} else {
					err = mig.Steps(-int(migrateSteps))
				}

				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}

				if migrateSteps != 0 {
					dumpMigrateVersion(mig)
				}

			case "goto":
				assertMigrateVersion(cmd) // or exit

				err := mig.Migrate(uint(migrateVersion))
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}

				dumpMigrateVersion(mig)

			case "force":
				assertMigrateVersion(cmd) // or exit

				err := mig.Force(int(migrateVersion))
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}

				dumpMigrateVersion(mig)

			case "version":
				dumpMigrateVersion(mig)

			default:
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: unsupported argument \"%s\"", args[0]))
				fmt.Fprintln(os.Stderr, cmd.UsageString())
				os.Exit(1)
			}
		},
	}
	migrateSteps   uint
	migrateVersion int
	migrateDbUri   string
)

func dumpMigrateVersion(mig *migrate.Migrate) {
	ver, dirty, err := mig.Version()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf("Schema version: %v\nIs dirty: %v\n", ver, dirty)
}

func getMigrateInstance(uri string) (*migrate.Migrate, error) {
	files := postgres.GetMigrations()
	source, err := httpfs.New(http.FS(files), "migrations")
	m, err := migrate.NewWithSourceInstance("httpfs", source, uri)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func assertMigrateVersion(cmd *cobra.Command) {
	if migrateVersion < 0 {
		fmt.Fprintln(os.Stderr, "Error: required flag(s) \"version\" not set")
		fmt.Fprintln(os.Stderr, cmd.UsageString())
		os.Exit(1)
	}
}

func init() {
	dbCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&migrateDbUri, "database", "d", "", "Database URI")
	migrateCmd.Flags().IntVarP(&migrateVersion, "version", "v", -1, "Database schema version")
	migrateCmd.Flags().UintVarP(&migrateSteps, "steps", "s", 0, "Positive integer of schema steps")
	migrateCmd.MarkFlagRequired("database")
}
