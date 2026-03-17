package tui

import (
	"github.com/gdamore/tcell/v2"
)

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         tcell.Key
	Rune        rune
	Description string
}

// KeyMap holds all keyboard shortcuts
var KeyMap = []KeyBinding{
	{tcell.KeyUp, 0, "Move cursor up"},
	{tcell.KeyDown, 0, "Move cursor down"},
	{tcell.KeyLeft, 0, "Move cursor left"},
	{tcell.KeyRight, 0, "Move cursor right"},
	{tcell.KeyPgUp, 0, "Page up"},
	{tcell.KeyPgDn, 0, "Page down"},
	{tcell.KeyHome, 0, "Go to first row"},
	{tcell.KeyEnd, 0, "Go to last row"},
	{0, 's', "Sort by column (cycle: none → asc → desc)"},
	{0, 'f', "Filter rows with expression"},
	{0, 'g', "Show graph/chart for column"},
	{0, 'e', "Export current view to file"},
	{0, 'r', "Reset (clear filters and sorts)"},
	{0, '?', "Show help"},
	{0, 'q', "Quit"},
	{tcell.KeyEsc, 0, "Cancel / Close dialog"},
}

// GetHelpText returns formatted help text
func GetHelpText() string {
	help := "ZeaView Keyboard Shortcuts\n\n"
	help += "Navigation:\n"
	help += "  ↑↓←→      Move cursor\n"
	help += "  PgUp/PgDn Page up/down\n"
	help += "  Home/End  First/last row\n\n"
	help += "Operations:\n"
	help += "  s         Sort by current column\n"
	help += "  f         Filter with expression\n"
	help += "  g         Show graph for column\n"
	help += "  e         Export to file\n"
	help += "  r         Reset filters/sorts\n\n"
	help += "Other:\n"
	help += "  ?         Show this help\n"
	help += "  q         Quit viewer\n"
	help += "  Esc       Cancel/close dialog\n"
	return help
}
