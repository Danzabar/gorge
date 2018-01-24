package cmd

import (
	"fmt"

	"github.com/Danzabar/gorge/engine"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init creates a new gorge config file",
	Long: `The configuration file enables you to import other config files and folders
and control the flow of the application. It may soon become a requirement to the engine.`,
	Run: func(cmd *cobra.Command, args []string) {
		engine.WriteConfig(&engine.GorgeSettings{}, engine.STANDARD_CONFIG)
		fmt.Println("Configuration created...")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
