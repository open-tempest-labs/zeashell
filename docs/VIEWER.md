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
| `d` | Show schema and metadata |
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

## Schema

Press `d` to display the schema and metadata for the current dataset.

### Schema Display

Shows comprehensive information about your data:
- **Column names and types**: Each column with its inferred type (string, int64, float64, bool, multi)
- **Null statistics**: Count and percentage of null values per column
- **Total rows and columns**: Dataset dimensions
- **Active filters**: Currently applied filter expressions
- **Active sorts**: Current sort column and direction

Example:
```
┌─ Schema & Metadata ────────────────────────────────────┐
│ Schema:                                                 │
│ -------                                                 │
│   id                      int64                         │
│   customer                string                        │
│   amount                  float64                       │
│   region                  string      (2 nulls, 16.7%)  │
│   date                    string                        │
│                                                         │
│ Total Rows: 12                                          │
│ Total Columns: 5                                        │
│                                                         │
│ Active Filter: amount > 1000                            │
│ Original Rows: 50                                       │
│                                                         │
│ Press ESC or d to close                                 │
└─────────────────────────────────────────────────────────┘
```

This is particularly useful for:
- Understanding data types before filtering
- Checking data quality (null counts)
- Verifying filter impact (original vs filtered row counts)
- Quick schema reference without leaving the viewer

Press `d` or `Esc` to close the schema view and return to the table.

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
- **Row**: Current cursor position / total rows (updates as you navigate)
- **Col**: Current column position / total columns
- **Sort**: Current sort column and direction (if any)
- **Filter**: Current filter expression (if any)

Example:
```
Row: 1,247/5,000 | Col: 3/4 | Filter: amount > 1000 | Sort: amount (desc)
```

The status bar updates dynamically as you navigate through the data, helping you track your position in large datasets.

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

The viewer supports scrolling through datasets of any size using `PgUp`/`PgDn`, `Home`/`End`, and arrow keys. All rows are accessible without limitations.

For extremely large datasets (millions of rows), you can pre-filter or aggregate before viewing for better performance:

```bash
# Pre-filter large data before viewing
zea load "huge-data/*.parquet" \
  | zea filter "amount > 1000" \
  | zea view -

# Or aggregate first for summary view
zea load "huge-data/*.parquet" \
  | zea group region --sum=amount \
  | zea view -
```

The status bar always shows your current position (e.g., `Row: 1,247/1,000,000`) so you know where you are in the dataset.

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

- Values longer than 50 characters are truncated with "..." in table view (press Enter to see full value)
- Graphs show up to 10 bins for histograms
- Bar charts show up to 20 unique values

The viewer handles datasets of any size with full scrolling support. For extremely large datasets (millions of rows), consider pre-filtering for better performance.

## Related Documentation

- [EXPRESSIONS.md](EXPRESSIONS.md) - Complete filter expression syntax
- [COMMANDS.md](COMMANDS.md) - All ZeaShell commands
- [examples/FILTER_SYNTAX.md](../examples/FILTER_SYNTAX.md) - Filter syntax examples and troubleshooting
