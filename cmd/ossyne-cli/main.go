package main

import (
	"fmt"
	"os"
	"ossyne/internal/cli"
	"ossyne/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
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
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				p := tea.NewProgram(tui.InitModel(), tea.WithAltScreen())
				if _, err := p.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
					os.Exit(1)
				}
			} else {
				cmd.Help()
			}
		},
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
	rootCmd.AddCommand(cli.NewProjectCmd())
	rootCmd.AddCommand(cli.NewTaskCmd())
	rootCmd.AddCommand(cli.NewRatingsCmd())
	rootCmd.AddCommand(cli.NewAdminCmd())
	rootCmd.AddCommand(cli.NewPaymentCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}