package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/rivo/tview"
)

// TableView wraps tview.Table with ZeaFrame-specific functionality
type TableView struct {
	*tview.Table
	frame          *zeaframe.ZeaFrame
	currentRow     int
	currentCol     int
	visibleRows    int
	visibleCols    int
	scrollOffsetRow int
	scrollOffsetCol int
}

// NewTableView creates a new table view for a ZeaFrame
func NewTableView(zf *zeaframe.ZeaFrame) *TableView {
	tv := &TableView{
		Table: tview.NewTable(),
		frame: zf,
	}

	tv.SetSelectable(true, true).
		SetFixed(1, 0). // Fix header row
		SetSeparator(tview.Borders.Vertical)

	tv.populate()

	return tv
}

// populate fills the table with data from the ZeaFrame
func (tv *TableView) populate() {
	tv.Clear()

	// Add header row
	for col := 0; col < len(tv.frame.Columns); col++ {
		cell := tview.NewTableCell(tv.frame.Columns[col].Name).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetExpansion(1)
		tv.SetCell(0, col, cell)
	}

	// Add all data rows (tview handles scrolling automatically)
	for row := 0; row < tv.frame.Rows; row++ {
		for col := 0; col < len(tv.frame.Columns); col++ {
			value := tv.formatValue(col, row)

			cell := tview.NewTableCell(value).
				SetExpansion(1)

			// Alternate row colors
			if row%2 == 0 {
				cell.SetBackgroundColor(tcell.ColorDarkSlateGray)
			}

			tv.SetCell(row+1, col, cell)
		}
	}
}

// formatValue formats a cell value for display
func (tv *TableView) formatValue(col, row int) string {
	column := tv.frame.Columns[col]

	if column.Nulls[row] {
		return "NULL"
	}

	value := column.Data[row]

	switch v := value.(type) {
	case string:
		if len(v) > 50 {
			return v[:47] + "..."
		}
		return v
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 50 {
			return str[:47] + "..."
		}
		return str
	}
}

// GetCurrentColumn returns the name of the currently selected column
func (tv *TableView) GetCurrentColumn() string {
	_, col := tv.GetSelection()
	if col >= 0 && col < len(tv.frame.Columns) {
		return tv.frame.Columns[col].Name
	}
	return ""
}

// GetCurrentColumnIndex returns the index of the currently selected column
func (tv *TableView) GetCurrentColumnIndex() int {
	_, col := tv.GetSelection()
	return col
}

// UpdateFrame updates the table with a new ZeaFrame
func (tv *TableView) UpdateFrame(zf *zeaframe.ZeaFrame) {
	tv.frame = zf
	tv.populate()
	tv.Select(1, 0) // Reset to first data row
}
