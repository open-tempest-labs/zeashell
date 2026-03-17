package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/rivo/tview"
)

// Viewer represents the TUI application state
type Viewer struct {
	app           *tview.Application
	pages         *tview.Pages
	table         *TableView
	originalFrame *zeaframe.ZeaFrame
	currentFrame  *zeaframe.ZeaFrame
	sortColumn    string
	sortAsc       bool
	filterExpr    string
	status        *tview.TextView
}

// RunViewFromSource launches the TUI viewer from a data source
func RunViewFromSource(source string, stdin io.Reader) error {
	// Check if we have a TTY
	if !isTerminal() {
		return fmt.Errorf("zea view requires a terminal (TTY)")
	}

	var zf *zeaframe.ZeaFrame
	var err error

	// Load data
	if source == "" || source == "-" {
		// Read from stdin
		zf, err = zeaframe.FromCSV(stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		// Load from file (supports glob patterns, directories, etc.)
		if zeaframe.IsGlobPattern(source) || zeaframe.IsDirectory(source) {
			opts := zeaframe.DefaultGlobOptions()
			files, err := zeaframe.GlobFiles(source, opts)
			if err != nil {
				return fmt.Errorf("failed to resolve files: %w", err)
			}
			zf, err = zeaframe.LoadMultipleFiles(files, opts)
			if err != nil {
				return fmt.Errorf("failed to load files: %w", err)
			}
		} else {
			zf, err = zeaframe.LoadAuto(source)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", source, err)
			}
		}
	}

	// Create and run viewer
	viewer := NewViewer(zf)
	return viewer.Run()
}

// NewViewer creates a new TUI viewer
func NewViewer(zf *zeaframe.ZeaFrame) *Viewer {
	app := tview.NewApplication()
	pages := tview.NewPages()

	v := &Viewer{
		app:           app,
		pages:         pages,
		originalFrame: zf,
		currentFrame:  zf,
	}

	// Create main layout
	v.setupUI()

	return v
}

// setupUI initializes the UI layout
func (v *Viewer) setupUI() {
	// Create table view
	v.table = NewTableView(v.currentFrame)
	v.table.SetBorder(true).SetTitle(" ZeaView ")

	// Create status bar
	v.status = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	v.updateStatus()

	// Create main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.table, 0, 1, true).
		AddItem(v.status, 1, 0, false)

	// Set input capture for keyboard shortcuts
	flex.SetInputCapture(v.handleInput)

	// Add to pages
	v.pages.AddPage("main", flex, true, true)

	v.app.SetRoot(v.pages, true)
}

// handleInput processes keyboard input
func (v *Viewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 's':
		v.handleSort()
		return nil
	case 'f':
		v.showFilterDialog()
		return nil
	case 'g':
		v.showGraph()
		return nil
	case 'e':
		v.showExportDialog()
		return nil
	case 'r':
		v.handleReset()
		return nil
	case '?':
		v.showHelp()
		return nil
	case 'q':
		v.app.Stop()
		return nil
	}

	return event
}

// updateStatus updates the status bar
func (v *Viewer) updateStatus() {
	status := fmt.Sprintf(" Rows: %d | Cols: %d", v.currentFrame.Rows, len(v.currentFrame.Columns))

	if v.sortColumn != "" {
		dir := "asc"
		if !v.sortAsc {
			dir = "desc"
		}
		status += fmt.Sprintf(" | Sort: %s (%s)", v.sortColumn, dir)
	}

	if v.filterExpr != "" {
		status += fmt.Sprintf(" | Filter: %s", v.filterExpr)
	}

	status += " | Press ? for help"

	v.status.SetText(status)
}

// Run starts the TUI application
func (v *Viewer) Run() error {
	return v.app.Run()
}

// isTerminal checks if stdin/stdout are connected to a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
