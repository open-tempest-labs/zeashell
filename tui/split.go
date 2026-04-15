package tui

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/gdamore/tcell/v2"
	"github.com/open-tempest-labs/zeashell/zeaframe"
	"github.com/rivo/tview"
)

// SplitPane describes one pane in a split view.
type SplitPane struct {
	Name    string
	Schema  *arrow.Schema
	Records []arrow.Record
}

// RunSplitViewFromArrow displays multiple Arrow tables in a split layout.
// orientation must be "top-bottom" (default) or "left-right".
// Each pane is fully independent — sort, filter, and scroll state are not shared.
// Tab / Shift-Tab cycle focus between panes. q quits from any pane.
func RunSplitViewFromArrow(panes []SplitPane, orientation string) error {
	if !isTerminal() {
		return fmt.Errorf("zea view requires a terminal (TTY)")
	}
	if len(panes) == 0 {
		return fmt.Errorf("split view: no panes provided")
	}
	if len(panes) == 1 {
		return RunViewFromArrow(panes[0].Schema, panes[0].Records, nil)
	}

	app := tview.NewApplication()

	viewers := make([]*Viewer, len(panes))
	for i, p := range panes {
		zf, err := zeaframe.FromArrow(p.Schema, p.Records)
		if err != nil {
			return fmt.Errorf("split view pane %q: %w", p.Name, err)
		}
		viewers[i] = newPaneViewer(app, zf, p.Name)
	}

	focused := 0
	setFocus := func(idx int) {
		// dim the outgoing pane border
		viewers[focused].table.SetBorderColor(tcell.ColorDefault)
		viewers[focused].table.SetTitle(fmt.Sprintf(" %s ", panes[focused].Name))
		focused = idx
		// highlight the incoming pane border
		viewers[focused].table.SetBorderColor(tcell.ColorYellow)
		viewers[focused].table.SetTitle(fmt.Sprintf(" [yellow]%s[-] ", panes[focused].Name))
		app.SetFocus(viewers[focused].table)
	}

	dir := tview.FlexRow
	if orientation == "left-right" {
		dir = tview.FlexColumn
	}
	outer := tview.NewFlex().SetDirection(dir)
	for _, v := range viewers {
		outer.AddItem(v.pages, 0, 1, false)
	}

	// App-level Tab / Shift-Tab to cycle panes; q to quit.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			setFocus((focused + 1) % len(viewers))
			return nil
		case tcell.KeyBacktab:
			setFocus((focused + len(viewers) - 1) % len(viewers))
			return nil
		}
		return event
	})

	app.SetRoot(outer, true)
	setFocus(0)
	return app.Run()
}

// newPaneViewer creates a Viewer that shares an existing tview.Application.
// It sets up its own table, status bar, and pages but does NOT call
// app.SetRoot — the caller is responsible for the root layout.
func newPaneViewer(app *tview.Application, zf *zeaframe.ZeaFrame, name string) *Viewer {
	pages := tview.NewPages()

	v := &Viewer{
		app:           app,
		pages:         pages,
		originalFrame: zf,
		currentFrame:  zf,
	}

	v.table = NewTableView(v.currentFrame)
	v.table.SetBorder(true).SetTitle(fmt.Sprintf(" %s ", name))

	v.table.SetSelectionChangedFunc(func(row, col int) {
		v.updateStatus()
	})

	v.status = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	v.updateStatus()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.table, 0, 1, true).
		AddItem(v.status, 1, 0, false)

	flex.SetInputCapture(v.handleInput)

	pages.AddPage("main", flex, true, true)

	return v
}
