package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zea",
	Short: "ZeaShell - DataFrame Shell for modern file formats",
	Long: `ZeaShell is a DataFrame shell that processes modern file formats with Unix pipe semantics.

ZeaShell brings integrated data exploration and transformation to the command line - combining
interactive discovery, powerful pipelines, and multi-valued data support for modern formats.`,
	Version: "0.4.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Record every invocation (except pluginize itself) to ~/.zea/history.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if len(os.Args) > 1 && os.Args[1] != "pluginize" {
			appendHistory(formatArgs(os.Args[1:]))
		}
	}

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(sortCmd)
	rootCmd.AddCommand(groupCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(joinCmd)
	rootCmd.AddCommand(pivotCmd)
	rootCmd.AddCommand(unpivotCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(sqlCmd)
	rootCmd.AddCommand(pluginizeCmd)

	// Load plugins from ~/.zea/plugins or $ZEA_PLUGINS
	loadPlugins()
}
