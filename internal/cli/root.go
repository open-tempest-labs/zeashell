package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zea",
	Short: "ZeaShell - DataFrame Shell for modern file formats",
	Long: `ZeaShell is a DataFrame shell that processes modern file formats with Unix pipe semantics.

ZeaShell brings integrated data exploration and transformation to the command line - combining
interactive discovery, powerful pipelines, and multi-valued data support for modern formats.`,
	Version: "0.2.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
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
}
