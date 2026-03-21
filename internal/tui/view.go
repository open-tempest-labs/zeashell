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

	// Update status bar when selection changes
	v.table.SetSelectionChangedFunc(func(row, col int) {
		v.updateStatus()
	})

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
	case 'd':
		v.showSchema()
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

	// Handle special keys
	switch event.Key() {
	case tcell.KeyEnter:
		v.showFullCellValue()
		return nil
	}

	return event
}

// updateStatus updates the status bar
func (v *Viewer) updateStatus() {
	// Get current cursor position
	row, col := v.table.GetSelection()

	// Row 0 is header, so data rows start at 1
	currentRow := row
	if currentRow > 0 {
		currentRow-- // Convert to 0-based data row index
	}

	// Build status with current position
	status := fmt.Sprintf(" Row: %d/%d | Col: %d/%d",
		currentRow+1, v.currentFrame.Rows,
		col+1, len(v.currentFrame.Columns))

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

// showFullCellValue displays the complete value of the selected cell in a modal
func (v *Viewer) showFullCellValue() {
	row, col := v.table.GetSelection()
	if row < 1 || col < 0 {
		// Row 0 is header, ignore
		return
	}

	// Get actual data row (subtract 1 for header)
	dataRow := row - 1
	if dataRow >= v.currentFrame.Rows || col >= len(v.currentFrame.Columns) {
		return
	}

	column := v.currentFrame.Columns[col]

	// Get the full value
	var fullValue string
	if column.Nulls[dataRow] {
		fullValue = "NULL"
	} else {
		fullValue = fmt.Sprintf("%v", column.Data[dataRow])
	}

	// Create a text view to display the value
	textView := tview.NewTextView().
		SetText(fullValue).
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)

	textView.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s [Row %d] - Press ESC or q to close ", column.Name, dataRow+1))

	// Handle input for the modal
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			v.pages.RemovePage("cell-view")
			return nil
		}
		return event
	})

	// Show the modal
	v.pages.AddPage("cell-view", textView, true, true)
}

// showSchema displays the schema and metadata for the current data
func (v *Viewer) showSchema() {
	var content string

	// Build schema information
	content += "Schema:\n"
	content += "-------\n"
	for _, col := range v.currentFrame.Columns {
		typeName := getTypeName(col.Type)

		// Count nulls in this column
		nullCount := 0
		for i := 0; i < v.currentFrame.Rows; i++ {
			if col.Nulls[i] {
				nullCount++
			}
		}

		nullPercent := 0.0
		if v.currentFrame.Rows > 0 {
			nullPercent = float64(nullCount) / float64(v.currentFrame.Rows) * 100.0
		}

		content += fmt.Sprintf("  %-25s %-10s", col.Name, typeName)
		if nullCount > 0 {
			content += fmt.Sprintf(" (%d nulls, %.1f%%)", nullCount, nullPercent)
		}
		content += "\n"
	}

	content += "\n"
	content += fmt.Sprintf("Total Rows: %d\n", v.currentFrame.Rows)
	content += fmt.Sprintf("Total Columns: %d\n", len(v.currentFrame.Columns))

	// Add filter/sort info if applicable
	if v.filterExpr != "" {
		content += fmt.Sprintf("\nActive Filter: %s\n", v.filterExpr)
		content += fmt.Sprintf("Original Rows: %d\n", v.originalFrame.Rows)
	}

	if v.sortColumn != "" {
		dir := "ascending"
		if !v.sortAsc {
			dir = "descending"
		}
		content += fmt.Sprintf("\nActive Sort: %s (%s)\n", v.sortColumn, dir)
	}

	// Create a text view to display the schema
	textView := tview.NewTextView().
		SetText(content).
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)

	textView.SetBorder(true).
		SetTitle(" Schema & Metadata - Press ESC or d to close ")

	// Handle input for the modal
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'd' || event.Rune() == 'q' {
			v.pages.RemovePage("schema-view")
			return nil
		}
		return event
	})

	// Show the modal
	v.pages.AddPage("schema-view", textView, true, true)
}

// getTypeName returns a human-readable type name
func getTypeName(colType zeaframe.ColumnType) string {
	switch colType {
	case zeaframe.StringType:
		return "string"
	case zeaframe.Int64Type:
		return "int64"
	case zeaframe.Float64Type:
		return "float64"
	case zeaframe.BoolType:
		return "bool"
	case zeaframe.MultiType:
		return "multi"
	default:
		return "unknown"
	}
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
