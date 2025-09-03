package main

import (
	"fmt"
	"os"
	"ossyne/internal/cli"
	"github.com/spf13/cobra"
)

var version = "0.0.1"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "osm",
		Short: "OSM is a CLI-first marketplace for open-source collaboration.",
		Long: `OSM (Open-Source Marketplace) is a command-line tool that connects
open-source project maintainers with contributors.

Find tasks, submit contributions, earn bounties, and build your
reputation in the open-source world, all from your terminal.`,
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of OSM",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("osm version %s\n", version)
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cli.NewUserCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}