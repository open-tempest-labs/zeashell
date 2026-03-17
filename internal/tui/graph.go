package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/open-tempest-labs/zeashell/internal/zeaframe"
	"github.com/rivo/tview"
)

// showGraph displays a graph/chart for the current column
func (v *Viewer) showGraph() {
	colIdx := v.table.GetCurrentColumnIndex()
	if colIdx < 0 || colIdx >= len(v.currentFrame.Columns) {
		return
	}

	column := v.currentFrame.Columns[colIdx]

	var graphText string
	switch column.Type {
	case zeaframe.Int64Type, zeaframe.Float64Type:
		graphText = v.generateNumericGraph(column)
	case zeaframe.StringType:
		graphText = v.generateCategoryGraph(column)
	default:
		graphText = "Graph not available for this column type"
	}

	textView := tview.NewTextView().
		SetText(graphText).
		SetDynamicColors(true)
	textView.SetBorder(true).SetTitle(fmt.Sprintf(" Graph: %s ", column.Name))

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'g' || event.Key() == tcell.KeyEsc {
			v.pages.RemovePage("graph")
			return nil
		}
		return event
	})

	// Full screen graph
	v.pages.AddPage("graph", textView, true, true)
}

// generateNumericGraph creates a histogram for numeric columns
func (v *Viewer) generateNumericGraph(column *zeaframe.Column) string {
	var values []float64

	// Collect non-null numeric values
	for i := 0; i < len(column.Data); i++ {
		if column.Nulls[i] {
			continue
		}

		var val float64
		switch v := column.Data[i].(type) {
		case int64:
			val = float64(v)
		case float64:
			val = v
		default:
			continue
		}
		values = append(values, val)
	}

	if len(values) == 0 {
		return "No data to display"
	}

	// Calculate statistics
	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))

	// Create histogram with 20 bins
	numBins := 20
	if len(values) < numBins {
		numBins = len(values)
	}

	bins := make([]int, numBins)
	binWidth := (max - min) / float64(numBins)

	if binWidth == 0 {
		// All values are the same
		bins[0] = len(values)
	} else {
		for _, val := range values {
			binIdx := int((val - min) / binWidth)
			if binIdx >= numBins {
				binIdx = numBins - 1
			}
			bins[binIdx]++
		}
	}

	// Find max bin count for scaling
	maxBin := 0
	for _, count := range bins {
		if count > maxBin {
			maxBin = count
		}
	}

	// Build graph
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Column: %s\n", column.Name))
	sb.WriteString(fmt.Sprintf("Count: %d | Min: %.2f | Max: %.2f | Avg: %.2f\n\n", len(values), min, max, avg))

	// Use Unicode block characters for bars
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	sb.WriteString("Distribution:\n")
	for i, count := range bins {
		binStart := min + float64(i)*binWidth
		binEnd := binStart + binWidth

		// Calculate bar height (0-8 blocks)
		height := 0
		if maxBin > 0 {
			height = int(float64(count) / float64(maxBin) * 8)
			if height > 8 {
				height = 8
			}
			if height == 0 && count > 0 {
				height = 1
			}
		}

		// Build bar
		bar := ""
		if height > 0 {
			bar = string(blocks[height-1])
			// Repeat for visual effect
			bar = strings.Repeat(bar, 3)
		}

		sb.WriteString(fmt.Sprintf("%8.2f - %8.2f: %s %d\n", binStart, binEnd, bar, count))
	}

	sb.WriteString("\nPress 'g' or Esc to close")

	return sb.String()
}

// generateCategoryGraph creates a bar chart for categorical columns
func (v *Viewer) generateCategoryGraph(column *zeaframe.Column) string {
	// Count frequency of each unique value
	freq := make(map[string]int)

	for i := 0; i < len(column.Data); i++ {
		if column.Nulls[i] {
			freq["NULL"]++
			continue
		}

		val := fmt.Sprintf("%v", column.Data[i])
		freq[val]++
	}

	// Sort by frequency (descending)
	type kv struct {
		key   string
		value int
	}

	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].value > sorted[j].value
	})

	// Show top 20
	if len(sorted) > 20 {
		sorted = sorted[:20]
	}

	// Find max frequency for scaling
	maxFreq := 0
	for _, kv := range sorted {
		if kv.value > maxFreq {
			maxFreq = kv.value
		}
	}

	// Build graph
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Column: %s\n", column.Name))
	sb.WriteString(fmt.Sprintf("Unique values: %d | Showing top %d\n\n", len(freq), len(sorted)))

	// Max bar width
	maxBarWidth := 50

	for _, kv := range sorted {
		// Truncate long keys
		key := kv.key
		if len(key) > 30 {
			key = key[:27] + "..."
		}

		// Calculate bar width
		barWidth := 0
		if maxFreq > 0 {
			barWidth = int(math.Ceil(float64(kv.value) / float64(maxFreq) * float64(maxBarWidth)))
		}

		bar := strings.Repeat("█", barWidth)

		sb.WriteString(fmt.Sprintf("%-30s | %s %d\n", key, bar, kv.value))
	}

	sb.WriteString("\nPress 'g' or Esc to close")

	return sb.String()
}
