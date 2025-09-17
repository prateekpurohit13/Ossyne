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

const ossyneBanner = `
  ___  ____ ______   ___   _ _____ 
 / _ \/ ___/ ___\ \ / / \ | | ____|
| | | \___ \___ \\ V /|  \| |  _|  
| |_| |___) |__) || | | |\  | |___ 
 \___/|____/____/ |_| |_| \_|_____|

`

func main() {
	var rootCmd = &cobra.Command{
		Use:   "osm",
		Short: "OSSYNE is a CLI-first marketplace for open-source collaboration.",
		Long: `OSSYNE (Open-Source Marketplace) is a command-line tool that connects
open-source project maintainers with contributors.

Find tasks, submit contributions, earn bounties, and build your
reputation in the open-source world, all from your terminal.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				fmt.Print(ossyneBanner)
				fmt.Println(cmd.Long)
				fmt.Println()

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

	rootCmd.SetHelpTemplate(ossyneBanner + "\n" + rootCmd.HelpTemplate())

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of OSSYNE",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ossyne version %s\n", version)
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cli.NewUserCmd())
	rootCmd.AddCommand(cli.NewProjectCmd())
	rootCmd.AddCommand(cli.NewTaskCmd())
	rootCmd.AddCommand(cli.NewRatingsCmd())
	rootCmd.AddCommand(cli.NewAdminCmd())
	rootCmd.AddCommand(cli.NewPaymentCmd())
	rootCmd.AddCommand(cli.NewAuthCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}