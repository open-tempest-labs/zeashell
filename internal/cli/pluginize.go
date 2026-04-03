package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

const historyMax = 50

var (
	pluginizeFlagLast bool
	pluginizeFlagCmd  string
)

var pluginizeCmd = &cobra.Command{
	Use:   "pluginize",
	Short: "Create a reusable plugin from command history",
	Long: `Create a reusable plugin script from command history.

Without flags, opens an interactive TUI history browser.
Select a command, provide a name and description, and it is saved
as an executable plugin to ~/.zea/plugins/.

Examples:
  zea pluginize
  zea pluginize --last
  zea pluginize --cmd="zea load ducks.parquet | zea sql 'SELECT * FROM t'"`,
	RunE: runPluginize,
}

func init() {
	pluginizeCmd.Flags().BoolVar(&pluginizeFlagLast, "last", false, "Promote last history entry to plugin")
	pluginizeCmd.Flags().StringVar(&pluginizeFlagCmd, "cmd", "", "Inline command string to save as plugin")
}

// getHistoryPath returns ~/.zea/history.
func getHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zea", "history"), nil
}

// appendHistory appends a command string to ~/.zea/history.
// Errors are silently ignored so they never block normal command execution.
func appendHistory(cmd string) {
	if strings.TrimSpace(cmd) == "" {
		return
	}
	path, err := getHistoryPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, cmd)
}

// shellHistoryPath returns the path to the shell history file, checking $HISTFILE,
// then $SHELL, then falling back to whichever of zsh/bash history files exists.
func shellHistoryPath() string {
	if h := os.Getenv("HISTFILE"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zsh_history")
	}
	if strings.Contains(shell, "bash") {
		return filepath.Join(home, ".bash_history")
	}
	// No $SHELL clue — try zsh first, then bash.
	zsh := filepath.Join(home, ".zsh_history")
	if _, err := os.Stat(zsh); err == nil {
		return zsh
	}
	return filepath.Join(home, ".bash_history")
}

// readShellHistory reads the last n entries from shell history that contain "zea",
// returned most-recent first. Handles both zsh (: timestamp:elapsed;cmd) and
// bash (plain text) formats.
func readShellHistory(n int) []string {
	path := shellHistoryPath()
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	// zsh history can contain multi-byte sequences; use a larger buffer.
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Strip zsh extended history prefix: ": timestamp:elapsed;command"
		if strings.HasPrefix(line, ": ") {
			if idx := strings.Index(line, ";"); idx != -1 {
				line = strings.TrimSpace(line[idx+1:])
			}
		}
		if line == "" {
			continue
		}
		// Keep only lines that actually invoke zea.
		if !containsZea(line) {
			continue
		}
		lines = append(lines, line)
	}

	// Take the last n, then reverse so index 0 = most recent.
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// containsZea reports whether a shell history line invokes the zea binary.
func containsZea(line string) bool {
	// Match "zea " at start, "| zea " mid-pipeline, or bare "zea" at end.
	return strings.HasPrefix(line, "zea ") ||
		strings.HasPrefix(line, "./zea ") ||
		strings.Contains(line, " zea ") ||
		strings.Contains(line, "|zea ") ||
		line == "zea" ||
		line == "./zea"
}

// formatArgs produces a shell-safe string from os.Args slices for history storage.
func formatArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

// shellQuote wraps s in single quotes if it contains shell metacharacters.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\"'\\|&;()<>{}$`!*?[") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func runPluginize(cmd *cobra.Command, args []string) error {
	if !pluginizeTTY() {
		return fmt.Errorf("pluginize requires a terminal (TTY)")
	}

	var selectedCmd string

	switch {
	case pluginizeFlagCmd != "":
		selectedCmd = pluginizeFlagCmd
	case pluginizeFlagLast:
		history := readShellHistory(1)
		if len(history) == 0 {
			return fmt.Errorf("no zea commands found in shell history")
		}
		selectedCmd = history[0]
	default:
		var err error
		selectedCmd, err = runHistoryBrowser()
		if err != nil {
			return err
		}
		if selectedCmd == "" {
			return nil // user quit without selecting
		}
	}

	return runMetadataForm(selectedCmd)
}

// runHistoryBrowser opens a tview table of recent commands and returns the one selected.
// Returns ("", nil) if the user quits without selecting.
func runHistoryBrowser() (string, error) {
	history := readShellHistory(historyMax)
	if len(history) == 0 {
		return "", fmt.Errorf("no zea commands found in shell history (%s)", shellHistoryPath())
	}

	app := tview.NewApplication()
	var selected string

	table := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)

	table.SetCell(0, 0,
		tview.NewTableCell("  #  ").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false))
	table.SetCell(0, 1,
		tview.NewTableCell("Command").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false))

	for i, entry := range history {
		idCell := tview.NewTableCell(fmt.Sprintf("  %d  ", i+1)).
			SetTextColor(tcell.ColorDarkGray)
		cmdCell := tview.NewTableCell(entry).SetExpansion(1)
		if i%2 == 0 {
			idCell.SetBackgroundColor(tcell.ColorDarkSlateGray)
			cmdCell.SetBackgroundColor(tcell.ColorDarkSlateGray)
		}
		table.SetCell(i+1, 0, idCell)
		table.SetCell(i+1, 1, cmdCell)
	}

	table.SetBorder(true).SetTitle(" Recent Commands ")

	status := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [yellow]↑↓[white] Navigate  [yellow]Enter[white] Select  [yellow]q/Esc[white] Quit")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(status, 1, 0, false)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc, event.Rune() == 'q':
			app.Stop()
			return nil
		case event.Key() == tcell.KeyEnter:
			row, _ := table.GetSelection()
			if row >= 1 && row <= len(history) {
				selected = history[row-1]
				app.Stop()
			}
			return nil
		}
		return event
	})

	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

// runMetadataForm collects plugin name/desc/args and saves the plugin.
func runMetadataForm(selectedCmd string) error {
	app := tview.NewApplication()
	var saveErr error
	cancelled := false

	// Suggest a name from the first subcommand word.
	suggestedName := ""
	if parts := strings.Fields(selectedCmd); len(parts) > 0 {
		suggestedName = parts[0]
	}

	form := tview.NewForm()
	form.SetBorder(false)
	form.AddInputField("Name", suggestedName, 40, nil, nil)
	form.AddInputField("Description", "", 60, nil, nil)
	form.AddInputField("Args hint", "", 40, nil, nil)

	form.AddButton("Save", func() {
		name := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		desc := strings.TrimSpace(form.GetFormItem(1).(*tview.InputField).GetText())
		argsHint := strings.TrimSpace(form.GetFormItem(2).(*tview.InputField).GetText())
		if name == "" {
			return // keep form open
		}
		pluginsDir, err := getPluginsDir()
		if err != nil {
			saveErr = err
			app.Stop()
			return
		}
		saveErr = savePlugin(pluginsDir, selectedCmd, name, desc, argsHint)
		app.Stop()
	})

	form.AddButton("Cancel", func() {
		cancelled = true
		app.Stop()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cancelled = true
			app.Stop()
			return nil
		}
		return event
	})

	cmdView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[yellow]Command:[white] %s", selectedCmd)).
		SetWrap(true)

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cmdView, 3, 0, false).
		AddItem(form, 0, 1, true)
	container.SetBorder(true).SetTitle(" New Plugin ")

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(container, 16, 1, true).
			AddItem(nil, 0, 1, false), 72, 1, true).
		AddItem(nil, 0, 1, false)

	app.SetRoot(modal, true)
	if err := app.Run(); err != nil {
		return err
	}
	if cancelled {
		return nil
	}
	return saveErr
}

// savePlugin writes an executable bash plugin script to pluginsDir/name.
func savePlugin(pluginsDir, cmd, name, desc, argsHint string) error {
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	scriptPath := filepath.Join(pluginsDir, name)

	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString(fmt.Sprintf("# @name %s\n", name))
	if desc != "" {
		sb.WriteString(fmt.Sprintf("# @desc %s\n", desc))
	}
	if argsHint != "" {
		sb.WriteString(fmt.Sprintf("# @args %s\n", argsHint))
	}
	sb.WriteString("\n")
	sb.WriteString(cmd + "\n")

	if err := os.WriteFile(scriptPath, []byte(sb.String()), 0755); err != nil {
		return fmt.Errorf("failed to write plugin %s: %w", scriptPath, err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Plugin saved: %s\n", scriptPath)
	fmt.Fprintf(os.Stderr, "Run with: zea run %s\n", name)
	return nil
}

// pluginizeTTY reports whether stdout is connected to a terminal.
func pluginizeTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
