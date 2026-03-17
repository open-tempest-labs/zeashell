package tui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/rivo/tview"
)

// handleSort cycles through sort states for the current column
func (v *Viewer) handleSort() {
	colName := v.table.GetCurrentColumn()
	if colName == "" {
		return
	}

	// Cycle: no sort → asc → desc → no sort
	if v.sortColumn != colName {
		// Start sorting this column ascending
		v.sortColumn = colName
		v.sortAsc = true
	} else if v.sortAsc {
		// Switch to descending
		v.sortAsc = false
	} else {
		// Clear sort
		v.sortColumn = ""
	}

	// Apply sort
	v.applyTransformations()
}

// handleReset clears all filters and sorts
func (v *Viewer) handleReset() {
	v.sortColumn = ""
	v.sortAsc = true
	v.filterExpr = ""
	v.currentFrame = v.originalFrame
	v.table.UpdateFrame(v.currentFrame)
	v.updateStatus()
}

// applyTransformations applies current sort and filter to the original frame
func (v *Viewer) applyTransformations() {
	result := v.originalFrame

	// Apply filter first
	if v.filterExpr != "" {
		expr, err := zeaframe.ParseExpression(v.filterExpr)
		if err != nil {
			v.showError(fmt.Sprintf("Filter parse error: %v", err))
			return
		}
		filtered, err := result.Filter(expr)
		if err != nil {
			v.showError(fmt.Sprintf("Filter error: %v", err))
			return
		}
		result = filtered
	}

	// Apply sort
	if v.sortColumn != "" {
		order := zeaframe.Ascending
		if !v.sortAsc {
			order = zeaframe.Descending
		}
		sorted, err := result.Sort(zeaframe.SortColumn{
			Name:  v.sortColumn,
			Order: order,
		})
		if err != nil {
			v.showError(fmt.Sprintf("Sort error: %v", err))
			return
		}
		result = sorted
	}

	v.currentFrame = result
	v.table.UpdateFrame(result)
	v.updateStatus()
}

// showFilterDialog shows a dialog to input filter expression
func (v *Viewer) showFilterDialog() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Filter Expression ")

	// Add input field with current filter
	form.AddInputField("Expression", v.filterExpr, 50, nil, nil)

	form.AddButton("Apply", func() {
		expr := form.GetFormItem(0).(*tview.InputField).GetText()
		v.filterExpr = expr
		v.applyTransformations()
		v.pages.RemovePage("filter")
	})

	form.AddButton("Clear", func() {
		v.filterExpr = ""
		v.applyTransformations()
		v.pages.RemovePage("filter")
	})

	form.AddButton("Cancel", func() {
		v.pages.RemovePage("filter")
	})

	// Center the form
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 7, 1, true).
			AddItem(nil, 0, 1, false), 80, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("filter", modal, true, true)
}

// showExportDialog shows a dialog to export the current view
func (v *Viewer) showExportDialog() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Export Data ")

	form.AddInputField("File path", "output.csv", 50, nil, nil)

	form.AddButton("Export", func() {
		path := form.GetFormItem(0).(*tview.InputField).GetText()
		err := v.exportToFile(path)
		v.pages.RemovePage("export")

		if err != nil {
			v.showError(fmt.Sprintf("Export failed: %v", err))
		} else {
			v.showMessage(fmt.Sprintf("Exported to %s", path))
		}
	})

	form.AddButton("Cancel", func() {
		v.pages.RemovePage("export")
	})

	// Center the form
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 6, 1, true).
			AddItem(nil, 0, 1, false), 70, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("export", modal, true, true)
}

// exportToFile exports the current frame to a file
func (v *Viewer) exportToFile(path string) error {
	// Create or truncate the file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write CSV
	return v.currentFrame.WriteCSV(file)
}

// showHelp displays the help overlay
func (v *Viewer) showHelp() {
	helpText := GetHelpText()

	textView := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true)
	textView.SetBorder(true).SetTitle(" Help ")

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		v.pages.RemovePage("help")
		return nil
	})

	// Center the help
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 20, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("help", modal, true, true)
}

// showError displays an error message
func (v *Viewer) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("error")
		})

	v.pages.AddPage("error", modal, true, true)
}

// showMessage displays an info message
func (v *Viewer) showMessage(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("message")
		})

	v.pages.AddPage("message", modal, true, true)
}
