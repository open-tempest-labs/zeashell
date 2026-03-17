# ZeaShell Interactive Viewer

The `zea view` command launches an interactive terminal UI for exploring data with keyboard-driven operations.

## Quick Start

```bash
# View a file
zea view sales.csv

# View partitioned data
zea view "sales/**/*.parquet"

# View from URL
zea view "https://example.com/data.csv"

# View pipeline output (use - for stdin)
zea load sales.csv | zea filter "amount > 100" | zea view -
```

## Features

- **Scrollable table view** with syntax highlighting
- **Sort by column** (ascending/descending/none)
- **Filter with expressions** using inline syntax help
- **Graph/chart view** (histograms for numeric, bar charts for categorical)
- **Export** filtered/sorted view to CSV
- **Full keyboard navigation**

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑↓←→` | Move cursor between cells |
| `PgUp` / `PgDn` | Page up / Page down |
| `Home` | First row |
| `End` | Last row |

### Operations

| Key | Action |
|-----|--------|
| `s` | Sort by current column (cycle: none → asc → desc → none) |
| `f` | Open filter expression dialog |
| `g` | Show graph/chart for current column |
| `e` | Export current view to CSV file |
| `r` | Reset all filters and sorts |
| `?` | Show help overlay |
| `q` | Quit viewer |

## Sort

Press `s` on any column to cycle through sort states:
1. **First press**: Sort ascending
2. **Second press**: Sort descending
3. **Third press**: Remove sort (back to original order)

The status bar shows the current sort state, e.g., `Sort: amount (desc)`.

## Filter

Press `f` to open the filter expression dialog.

### Filter Dialog

The dialog shows:
- **Examples**: Quick reference for common patterns
- **Operators**: Available comparison and logical operators
- **Input field**: Enter your filter expression
- **Buttons**: Apply, Clear, Cancel

### Filter Syntax

```bash
# Simple comparisons
amount > 1000
region = 'West'

# Logical combinations (AND/OR must be uppercase)
amount > 1000 AND region = 'West'
amount < 500 OR amount > 3000

# Path-based columns (from JSON/XML)
address.city = 'SF'
service.role = 'WEBHDFS'

# Pattern matching
name CONTAINS '*.webshell.*'
```

**Important:**
- Column names: NO quotes
- String values: Single quotes `'value'`
- Numbers: NO quotes
- Logical operators: UPPERCASE (`AND`, `OR`)

See [EXPRESSIONS.md](EXPRESSIONS.md) and [examples/FILTER_SYNTAX.md](../examples/FILTER_SYNTAX.md) for complete syntax guide.

## Graph

Press `g` on any column to show a graph/chart.

### Numeric Columns (Histogram)

For numeric columns (int64, float64), shows a histogram with:
- **Distribution**: Number of values in each bin
- **Statistics**: Count, Min, Max, Average
- **Visual bars**: Unicode block characters sized by frequency

Example:
```
┌─ Graph: amount ───────────────────────────────────────────┐
│ Column: amount                                            │
│ Count: 12 | Min: 600.00 | Max: 4200.00 | Avg: 1783.33    │
│                                                           │
│ Distribution:                                             │
│      600 -      960: ███ 4                               │
│      960 -     1320: ▄▄▄ 2                               │
│     1320 -     1680: ▂▂▂ 1                               │
│     1680 -     2040: ▂▂▂ 1                               │
│     2040 -     2400: ▂▂▂ 1                               │
│                                                           │
│ Press 'g' or Esc to close                                 │
└───────────────────────────────────────────────────────────┘
```

### Categorical Columns (Bar Chart)

For string columns, shows a bar chart with:
- **Unique values**: Number of distinct values
- **Value counts**: Horizontal bars for each value
- **Visual bars**: Unicode block characters sized by count

Example:
```
┌─ Graph: region ───────────────────────────────────────────┐
│ Column: region                                            │
│ Unique values: 4                                          │
│                                                           │
│ West  | ████████████████████████████████████████ 4       │
│ East  | ██████████████████████████████ 3                 │
│ South | ██████████████████████████████ 3                 │
│ North | ████████████████████ 2                           │
│                                                           │
│ Press 'g' or Esc to close                                 │
└───────────────────────────────────────────────────────────┘
```

Press `g` or `Esc` to close the graph and return to table view.

## Export

Press `e` to export the current view (with any filters/sorts applied) to a CSV file.

### Export Dialog

1. Enter the file path (default: `output.csv`)
2. Press Enter or click **Export** to save
3. Success/error message appears

The export includes:
- All columns currently visible
- Only rows that pass the current filter (if any)
- Rows in the current sort order (if any)

## Reset

Press `r` to clear all filters and sorts, returning to the original data view.

This is useful when you've applied multiple transformations and want to start fresh.

## Help

Press `?` to show the help overlay with all keyboard shortcuts.

Press any key to close the help and return to the table view.

## Status Bar

The bottom status bar shows:
- **Rows**: Total number of rows in current view
- **Cols**: Total number of columns
- **Sort**: Current sort column and direction (if any)
- **Filter**: Current filter expression (if any)

Example:
```
Rows: 7 | Cols: 4 | Filter: amount > 1000 | Sort: amount (desc)
```

## Common Workflows

### Data Exploration

```bash
# Open viewer
zea view sales.csv

# Navigate with arrow keys to explore columns
# Press 's' on different columns to see sorted views
# Press 'g' on numeric columns to see distributions
# Press 'g' on categorical columns to see value counts
```

### Finding Outliers

```bash
# Open viewer
zea view sales.csv

# Sort by amount column (press 's')
# Look at top/bottom values
# Press 'g' to see histogram of distribution
```

### Interactive Filtering

```bash
# Open viewer
zea view sales.csv

# Press 'f' to open filter
# Enter: amount > 1000 AND region = 'West'
# See filtered results immediately
# Press 'e' to export filtered data
```

### Multi-Step Analysis

```bash
# Open viewer
zea view sales.csv

# Step 1: Filter by region
# Press 'f', enter: region = 'West'

# Step 2: Sort by amount
# Navigate to amount column, press 's' twice for desc

# Step 3: View distribution
# Press 'g' to see histogram

# Step 4: Export top performers
# Press 'e', save as: west_top_sales.csv
```

### Exploring Partitioned Data

```bash
# Load all partitions interactively
zea view "sales/date=*/*.parquet"

# Filter to specific date range
# Press 'f', enter: date >= '2026-03-01'

# Group analysis
# Press 'g' on region column to see distribution across regions
```

## Tips

1. **Start with sorting** - Quickly identify min/max values by sorting
2. **Use graphs for distributions** - Press 'g' to understand data shape
3. **Test filters interactively** - See results immediately without writing pipelines
4. **Export for further analysis** - Use 'e' to save filtered subsets
5. **Reset when confused** - Press 'r' to start over

## Advanced Usage

### View Pipeline Output

```bash
# Filter first, then view
zea load sales.csv | zea filter "amount > 100" | zea view -

# Join and view
zea join customers.csv orders.csv --on=cust_id | zea view -
```

### Large Datasets

For large datasets (millions of rows), the viewer shows the first 1000 rows. For full analysis, use command-line pipelines:

```bash
# Use pipelines for large data
zea load "huge-data/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount \
  | zea view -
```

The aggregated result will be small enough to explore interactively.

## Troubleshooting

### Filter Not Working

**Problem**: Filter shows no results or error message

**Common mistakes:**
1. Forgot quotes around strings: `region = 'West'` not `region = West`
2. Wrong case for logical operators: `AND` not `and`
3. Quoted numbers: `amount > 1000` not `amount > '1000'`
4. Column name typos: Check exact names in table header

See [examples/FILTER_SYNTAX.md](../examples/FILTER_SYNTAX.md) for complete troubleshooting guide.

### Graph Not Showing

**Problem**: Graph appears empty or shows error

**Causes:**
- All values are NULL
- Column is empty after filtering
- Data type mismatch

**Solution**: Press 'r' to reset filters and try again.

### Performance Issues

**Problem**: Viewer is slow with large files

**Solution**: Pre-filter data before viewing:
```bash
# Instead of:
zea view huge.csv

# Do:
zea load huge.csv | zea filter "relevant_condition" | zea view -
```

## Color Scheme

The viewer uses terminal colors for readability:
- **Header**: Yellow, bold
- **Current cell**: Highlighted with inverse colors
- **Status bar**: Info with active filters/sorts highlighted
- **Dialogs**: Bordered boxes with button highlights
- **Charts**: Unicode block characters (`▁▂▃▄▅▆▇█`)

## Limitations

- Table view shows first 1000 rows (filters and sorts apply to all data)
- Values longer than 50 characters are truncated with "..."
- Graphs show up to 10 bins for histograms
- Bar charts show up to 20 unique values

For full analysis of large datasets, use command-line pipelines and view aggregated results.

## Related Documentation

- [EXPRESSIONS.md](EXPRESSIONS.md) - Complete filter expression syntax
- [COMMANDS.md](COMMANDS.md) - All ZeaShell commands
- [examples/FILTER_SYNTAX.md](../examples/FILTER_SYNTAX.md) - Filter syntax examples and troubleshooting
