# ZeaShell TUI Viewer - Visual Mockup

## Main Table View

```
┌─ ZeaView ─────────────────────────────────────────────────────────────┐
│ customer │ region │ amount │ product      │                           │
├──────────┼────────┼────────┼──────────────┼───────────────────────────┤
│ Alice    │ West   │ 1500   │ Laptop       │                           │
│ Bob      │ East   │ 800    │ Monitor      │                           │
│ Charlie  │ West   │ 2200   │ Workstation  │                           │
│ Diana    │ South  │ 950    │ Keyboard     │                           │
│ Eve      │ North  │ 1200   │ Desktop      │                           │
│ Frank    │ West   │ 3500   │ Server       │ ◄── Current row           │
│ Grace    │ East   │ 600    │ Mouse        │                           │
│ Henry    │ South  │ 1800   │ Laptop       │                           │
│ Ivy      │ West   │ 2800   │ Workstation  │                           │
│ Jack     │ East   │ 1100   │ Monitor      │                           │
│ Kelly    │ North  │ 4200   │ Server       │                           │
│ Leo      │ South  │ 750    │ Keyboard     │                           │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ Rows: 12 | Cols: 4 | Sort: amount (desc) | Press ? for help          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Filter Dialog

```
┌─ ZeaView ─────────────────────────────────────────────────────────────┐
│                                                                         │
│                    ┌─ Filter Expression ─────────────┐                 │
│                    │                                  │                 │
│                    │ Expression: amount > 1000        │                 │
│                    │                                  │                 │
│                    │  [Apply]  [Clear]  [Cancel]     │                 │
│                    └──────────────────────────────────┘                 │
│                                                                         │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ Rows: 12 | Cols: 4 | Filter: amount > 1000 | Press ? for help         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Graph View - Numeric Column (Histogram)

```
┌─ Graph: amount ───────────────────────────────────────────────────────┐
│ Column: amount                                                         │
│ Count: 12 | Min: 600.00 | Max: 4200.00 | Avg: 1783.33                 │
│                                                                         │
│ Distribution:                                                           │
│      600 -      960: ███ 4                                             │
│      960 -     1320: ▄▄▄ 2                                             │
│     1320 -     1680: ▂▂▂ 1                                             │
│     1680 -     2040: ▂▂▂ 1                                             │
│     2040 -     2400: ▂▂▂ 1                                             │
│     2400 -     2760:  0                                                │
│     2760 -     3120: ▂▂▂ 1                                             │
│     3120 -     3480:  0                                                │
│     3480 -     3840: ▂▂▂ 1                                             │
│     3840 -     4200: ▂▂▂ 1                                             │
│                                                                         │
│ Press 'g' or Esc to close                                              │
└─────────────────────────────────────────────────────────────────────────┘
```

## Graph View - Categorical Column (Bar Chart)

```
┌─ Graph: region ───────────────────────────────────────────────────────┐
│ Column: region                                                         │
│ Unique values: 4                                                       │
│                                                                         │
│ West                 | ████████████████████████████████████████ 4     │
│ East                 | ██████████████████████████████ 3               │
│ South                | ██████████████████████████████ 3               │
│ North                | ████████████████████ 2                         │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
│ Press 'g' or Esc to close                                              │
└─────────────────────────────────────────────────────────────────────────┘
```

## Export Dialog

```
┌─ ZeaView ─────────────────────────────────────────────────────────────┐
│                                                                         │
│                       ┌─ Export Data ─────────────┐                    │
│                       │                            │                    │
│                       │ File path: output.csv      │                    │
│                       │                            │                    │
│                       │   [Export]   [Cancel]      │                    │
│                       └────────────────────────────┘                    │
│                                                                         │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ Rows: 7 | Cols: 4 | Filter: amount > 1000 | Press ? for help          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Help Overlay

```
┌─ ZeaView ─────────────────────────────────────────────────────────────┐
│                                                                         │
│                  ┌─ Help ────────────────────────────┐                 │
│                  │ ZeaView Keyboard Shortcuts        │                 │
│                  │                                    │                 │
│                  │ Navigation:                        │                 │
│                  │   ↑↓←→      Move cursor            │                 │
│                  │   PgUp/PgDn Page up/down           │                 │
│                  │   Home/End  First/last row         │                 │
│                  │                                    │                 │
│                  │ Operations:                        │                 │
│                  │   s         Sort by column         │                 │
│                  │   f         Filter with expression │                 │
│                  │   g         Show graph for column  │                 │
│                  │   e         Export to file         │                 │
│                  │   r         Reset filters/sorts    │                 │
│                  │                                    │                 │
│                  │ Other:                             │                 │
│                  │   ?         Show this help         │                 │
│                  │   q         Quit viewer            │                 │
│                  │   Esc       Cancel/close dialog    │                 │
│                  └────────────────────────────────────┘                 │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ Rows: 12 | Cols: 4 | Press ? for help                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## Workflow Example

### Step 1: Open viewer
```bash
$ zea view testdata/sales-partitioned/date=*/*.csv
```

### Step 2: Navigate and sort
- Use arrow keys to move to "amount" column
- Press `s` to sort ascending
- Press `s` again to sort descending

### Step 3: Filter data
- Press `f` to open filter dialog
- Type: `amount > 1000`
- Press Enter or click Apply

### Step 4: View graph
- Navigate to "region" column
- Press `g` to see bar chart of sales by region

### Step 5: Export filtered results
- Press `e` to open export dialog
- Enter filename: `high_value_sales.csv`
- Press Enter to export

### Step 6: Reset and explore more
- Press `r` to clear all filters/sorts
- Navigate to "product" column
- Press `g` to see product distribution

### Step 7: Quit
- Press `q` to exit

## Color Scheme (in actual TUI)

- **Header**: Yellow text, centered
- **Data rows**: Alternating dark gray / default background
- **Current cell**: Highlighted with inverse colors
- **Status bar**: Info text with active filters/sorts highlighted
- **Dialogs**: Bordered boxes with buttons
- **Charts**: Unicode block characters with bars
