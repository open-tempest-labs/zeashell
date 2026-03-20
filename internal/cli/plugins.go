package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// pluginMetadata holds parsed metadata from a plugin script
type pluginMetadata struct {
	name        string
	description string
	args        string
	path        string
}

// getPluginsDir returns the directory where plugins are stored
func getPluginsDir() (string, error) {
	pluginsDir := os.Getenv("ZEA_PLUGINS")
	if pluginsDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		pluginsDir = filepath.Join(homeDir, ".zea", "plugins")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(pluginsDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve plugins directory: %w", err)
	}

	return absPath, nil
}

// ensurePluginsDir creates the plugins directory if it doesn't exist
func ensurePluginsDir(pluginsDir string) error {
	info, err := os.Stat(pluginsDir)
	if os.IsNotExist(err) {
		// Create directory with 0755 permissions
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugins directory: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check plugins directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("plugins path exists but is not a directory: %s", pluginsDir)
	}
	return nil
}

// parsePluginMetadata reads the first 20 lines of a script and extracts metadata
func parsePluginMetadata(scriptPath string, filename string) (*pluginMetadata, error) {
	file, err := os.Open(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin script: %w", err)
	}
	defer file.Close()

	metadata := &pluginMetadata{
		name:        filename, // Default to filename
		description: "Plugin command",
		path:        scriptPath,
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 20 {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Skip empty lines and non-comment lines
		if !strings.HasPrefix(line, "#") {
			continue
		}

		// Remove leading '#' and trim
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))

		// Parse metadata directives
		if strings.HasPrefix(line, "@name ") {
			metadata.name = strings.TrimSpace(strings.TrimPrefix(line, "@name "))
		} else if strings.HasPrefix(line, "@desc ") {
			metadata.description = strings.TrimSpace(strings.TrimPrefix(line, "@desc "))
		} else if strings.HasPrefix(line, "@args ") {
			metadata.args = strings.TrimSpace(strings.TrimPrefix(line, "@args "))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading plugin script: %w", err)
	}

	return metadata, nil
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	// Check if it has executable permission
	// On Unix systems, check owner execute bit
	return info.Mode()&0111 != 0
}

// discoverPlugins scans the plugins directory and returns metadata for all valid plugins
func discoverPlugins(pluginsDir string) ([]*pluginMetadata, error) {
	var plugins []*pluginMetadata

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(pluginsDir, entry.Name())

		// Check if file is executable
		if !isExecutable(fullPath) {
			continue
		}

		// Parse metadata from the script
		metadata, err := parsePluginMetadata(fullPath, entry.Name())
		if err != nil {
			// Log error but continue with other plugins
			fmt.Fprintf(os.Stderr, "Warning: failed to parse plugin %s: %v\n", entry.Name(), err)
			continue
		}

		plugins = append(plugins, metadata)
	}

	return plugins, nil
}

// createPluginCommand creates a Cobra command for a plugin
func createPluginCommand(metadata *pluginMetadata) *cobra.Command {
	use := metadata.name
	if metadata.args != "" {
		use = fmt.Sprintf("%s %s", metadata.name, metadata.args)
	}

	cmd := &cobra.Command{
		Use:                use,
		Short:              metadata.description,
		Long:               metadata.description,
		DisableFlagParsing: true, // Pass all args to plugin
		Run: func(cmd *cobra.Command, args []string) {
			// Execute the plugin script with provided arguments
			execCmd := exec.Command(metadata.path, args...)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Plugin execution failed: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

// loadPlugins discovers and registers all plugins as subcommands
func loadPlugins() {
	// Get plugins directory
	pluginsDir, err := getPluginsDir()
	if err != nil {
		// Silently skip if we can't determine plugins directory
		if os.Getenv("ZEA_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "Debug: failed to get plugins directory: %v\n", err)
		}
		return
	}

	// Ensure plugins directory exists
	if err := ensurePluginsDir(pluginsDir); err != nil {
		// Silently skip if we can't create the directory
		if os.Getenv("ZEA_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "Debug: failed to ensure plugins directory: %v\n", err)
		}
		return
	}

	// Discover plugins
	plugins, err := discoverPlugins(pluginsDir)
	if err != nil {
		// Silently skip if we can't read the directory
		if os.Getenv("ZEA_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "Debug: failed to discover plugins: %v\n", err)
		}
		return
	}

	// Log plugin discovery in debug mode
	if os.Getenv("ZEA_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "Debug: discovered %d plugin(s) in %s\n", len(plugins), pluginsDir)
	}

	// Create a parent "run" command if there are plugins
	if len(plugins) > 0 {
		runCmd := &cobra.Command{
			Use:   "run",
			Short: "Run plugin commands",
			Long:  fmt.Sprintf("Run custom plugin commands from %s", pluginsDir),
		}

		// Register each plugin as a subcommand of "run"
		for _, metadata := range plugins {
			pluginCmd := createPluginCommand(metadata)
			runCmd.AddCommand(pluginCmd)

			if os.Getenv("ZEA_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "Debug: registered plugin '%s'\n", metadata.name)
			}
		}

		// Add the "run" command to the root command
		rootCmd.AddCommand(runCmd)
	}
}
