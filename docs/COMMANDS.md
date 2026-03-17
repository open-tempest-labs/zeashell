# ZeaShell Commands Reference

Complete reference for all ZeaShell commands.

## Command Overview

| Command | Description |
|---------|-------------|
| `zea load` | Load data from files, URLs, patterns, or directories |
| `zea select` | Select specific columns |
| `zea filter` | Filter rows with expressions |
| `zea sort` | Sort by one or more columns |
| `zea group` | Group by and aggregate |
| `zea join` | Join two DataFrames |
| `zea pivot` | Transform long to wide format |
| `zea unpivot` | Transform wide to long format |
| `zea view` | Interactive terminal UI viewer |
| `zea describe` | Show schema and preview |
| `zea store` | Write data to file |

---

## `zea load [file|url|pattern|directory]`

Load files, glob patterns, directories, or URLs and output to stdout. Format is auto-detected from file extension.

### Supported Sources

- **Single files**: `sales.csv`, `/path/to/data.json`
- **Glob patterns**: `*.csv`, `sales/**/*.parquet`
- **Directories**: `sales/` (recursive by default)
- **Multiple files**: `file1.csv,file2.csv` (comma-separated)
- **HTTP/HTTPS URLs**: `http://example.com/data.csv`
- **stdin**: `cat sales.csv | zea load`

### Multi-File Loading

When loading multiple files:
- Files are loaded in parallel for performance (default: 8 workers)
- Schema is inferred from first 3 files
- All files are unioned into a single table
- Missing columns are filled with NULLs (schema evolution)

### Flags

- `--max-files=N` - Limit number of files to load (default: unlimited)
- `--parallel=N` - Number of parallel workers (default: 8)
- `--format=fmt` - Filter by format (csv, parquet, json, etc.)
- `--schema-preview` - Show inferred schema without loading data

### Examples

```bash
# Single files
zea load sales.csv
zea load data.parquet
zea load https://example.com/data.csv

# Glob patterns
zea load "*.csv"
zea load "sales/date=*.parquet"
zea load "sales/date=2026-03-*/*.csv"

# Directories (recursive)
zea load "sales/"
zea load "sales/**/*.parquet"

# Multiple files
zea load "file1.csv,file2.csv"
zea load "data/*.csv,archive/*.csv"

# Schema preview
zea load "sales/" --schema-preview

# Performance tuning
zea load "sales/" --parallel=16 --max-files=100

# stdin
cat sales.csv | zea load

# Partitioned data workflows
zea load "sales/date=2026-03-*/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount
```

See [PARTITIONED_DATA.md](PARTITIONED_DATA.md) for detailed information on glob patterns and partitioned data.

---

## `zea select [columns]`

Select (project) specific columns from the DataFrame.

### Syntax

```bash
zea select column1,column2,column3
```

Columns are comma-separated, no spaces.

### Examples

```bash
zea load sales.csv | zea select customer,amount
zea load data.csv | zea select region,product,date
zea load data.json | zea select address.city,address.state
```

---

## `zea filter [expression]`

Filter rows based on boolean expressions.

See [EXPRESSIONS.md](EXPRESSIONS.md) for complete expression language reference.

### Quick Reference

**Operators:**
- Comparison: `=`, `!=`, `>`, `>=`, `<`, `<=`
- Logical: `AND`, `OR`
- Array/Pattern: `CONTAINS`

**Examples:**

```bash
# Simple comparisons
zea load sales.csv | zea filter "amount > 100"
zea load sales.csv | zea filter "region = 'west'"

# Logical combinations
zea load sales.csv | zea filter "amount > 100 AND region = 'west'"
zea load sales.csv | zea filter "amount < 50 OR amount > 1000"

# Path-based columns (JSON/XML)
zea load data.json | zea filter "address.city = 'SF'"
zea load topology.xml | zea filter "service.role = 'WEBHDFS'"

# Pattern matching
zea load data.csv | zea filter "name CONTAINS '*.webshell.*'"
```

---

## `zea sort [columns]`

Sort rows by one or more columns.

### Syntax

**Column format**: `column[:asc|:desc]`
- `column` - Sort ascending (default)
- `column:asc` - Sort ascending (explicit)
- `column:desc` - Sort descending

**Multiple columns** are applied in order (stable sort).

### Examples

```bash
# Single column
zea load sales.csv | zea sort amount
zea load sales.csv | zea sort amount:desc

# Multiple columns
zea load sales.csv | zea sort region,amount
zea load sales.csv | zea sort region,amount:desc

# In pipelines
zea load sales.csv \
  | zea filter "amount > 100" \
  | zea sort region,amount:desc \
  | zea select region,customer,amount
```

---

## `zea group [columns] [--agg=column]`

Group by columns and perform aggregations.

### Aggregations

- `--sum=column` - Sum of values
- `--avg=column` - Average of values
- `--min=column` - Minimum value
- `--max=column` - Maximum value
- `--count=column` - Count of rows (use `1` for row count)

Multiple aggregations can be specified.

### Examples

```bash
# Single aggregation
zea load sales.csv | zea group region --sum=amount

# Multiple aggregations
zea load sales.csv | zea group product --sum=amount --count=1

# Multiple grouping columns
zea load sales.csv | zea group region,product --sum=amount

# Complete analysis
zea load sales.csv \
  | zea group customer --sum=amount --count=1 --avg=amount
```

---

## `zea join [left-source] [right-source]`

Join two DataFrames on one or more key columns.

### Join Types

- `inner` - Only rows with matches in both datasets (default)
- `left` - All left rows, NULLs for unmatched right
- `right` - All right rows, NULLs for unmatched left
- `full` - All rows from both, NULLs where no match

### Flags

- `--on=column[,column2,...]` - Join key column(s) (required)
- `--type=inner|left|right|full` - Join type (default: inner)

### Column Collisions

Column name collisions are resolved by adding `_right` suffix to right-side columns.

### Examples

```bash
# Inner join on single key
zea join customers.csv orders.csv --on=cust_id

# Left join on multiple keys
zea join customers.csv orders.csv --on=id,date --type=left

# Join stdin with file
zea load customers.csv | zea join orders.csv --on=cust_id

# Find unmatched rows
zea join customers.csv orders.csv --on=cust_id --type=left \
  | zea filter "order_id IS NULL"

# Complex pipeline
zea load sales.csv \
  | zea filter "amount > 100" \
  | zea join products.csv --on=product_id \
  | zea select customer,product.name,amount \
  | zea store enriched.parquet
```

---

## `zea pivot`

Transform long format data to wide format.

### Flags

- `--index=column[,column2,...]` - Index column(s) to group by (required)
- `--column=column` - Column whose values become new column names (required)
- `--values=column` - Column whose values populate the new columns (required)

### Example Transformation

```
Input (long):                Output (wide):
date,region,amount          date,west,east
2026-01-01,west,100         2026-01-01,100,50
2026-01-01,east,50          2026-01-02,70,
2026-01-02,west,70
```

### Examples

```bash
# Simple pivot
zea load sales_long.csv \
  | zea pivot --index=date --column=region --values=amount

# Multiple index columns
zea load data.csv \
  | zea pivot --index=year,month --column=category --values=sales

# Pivot in pipeline
zea load sales.csv \
  | zea filter "amount > 0" \
  | zea pivot --index=date --column=product --values=amount \
  | zea store sales_wide.parquet
```

---

## `zea unpivot`

Transform wide format data to long format.

### Flags

- `--id=column[,column2,...]` - ID column(s) to preserve (optional)
- `--values=column[,column2,...]` - Columns to unpivot into rows (required)
- `--name=column` - Name for column containing original column names (default: variable)
- `--value=column` - Name for column containing values (default: value)

### Example Transformation

```
Input (wide):                 Output (long):
date,west,east               date,region,amount
2026-01-01,100,50            2026-01-01,west,100
2026-01-02,70,               2026-01-01,east,50
                             2026-01-02,west,70
```

### Examples

```bash
# Simple unpivot
zea load sales_wide.csv \
  | zea unpivot --id=date --values=west,east --name=region --value=amount

# Multiple ID columns
zea load data.csv \
  | zea unpivot --id=year,month --values=q1,q2,q3,q4 --name=quarter --value=sales

# Unpivot in pipeline
zea load sales_wide.csv \
  | zea unpivot --id=date --values=west,east,north,south --name=region --value=amount \
  | zea filter "amount > 100" \
  | zea store sales_long.csv
```

---

## `zea view [source]`

Interactive terminal UI for data exploration.

See [VIEWER.md](VIEWER.md) for complete viewer documentation.

### Quick Start

```bash
# View single file
zea view sales.csv

# View partitioned data
zea view "sales/**/*.parquet"

# View from URL
zea view "https://example.com/data.csv"
```

### Keyboard Shortcuts

- `↑↓←→` - Navigate cells
- `s` - Sort by current column
- `f` - Filter with expression
- `g` - Show graph/chart
- `e` - Export to CSV
- `r` - Reset filters and sorts
- `?` - Show help
- `q` - Quit

---

## `zea describe`

Show schema and preview of the DataFrame.

### Examples

```bash
zea load sales.csv | zea describe
zea load data.csv | zea filter "amount > 100" | zea describe
```

### Output

Shows:
- Column names and types
- Row count
- Preview of first few rows

---

## `zea store [file]`

Store data to a file (or stdout). Format is auto-detected from file extension.

### Supported Formats

- `.csv` - CSV format
- `.tsv` - TSV format
- `.json` - JSON array of objects
- `.jsonl` - JSON Lines (one object per line)
- `.xml` - XML format
- `.parquet` - Apache Parquet format

### Examples

```bash
# Store to different formats
zea load sales.csv | zea filter "amount > 100" | zea store output.csv
zea load data.csv | zea select customer,total | zea store summary.tsv
zea load data.csv | zea filter "amount > 100" | zea store filtered.json
zea load events.csv | zea filter "status = 'active'" | zea store events.jsonl
zea load topology.xml | zea filter "service.role != ''" | zea store services.xml
zea load sales.csv | zea filter "amount > 1000" | zea store high_value.parquet

# Store to stdout (default format: CSV)
zea load sales.csv | zea filter "amount > 100" | zea store
```

---

## Piping Between Commands

ZeaShell commands can be chained together with Unix pipes:

```bash
# Basic pipeline
zea load sales.csv | zea filter "amount > 100" | zea select customer,amount

# Complex analysis
zea load sales.csv \
  | zea filter "amount > 100 AND region = 'west'" \
  | zea group customer --sum=amount --count=1 \
  | zea filter "amount_sum > 1000" \
  | zea store west_vip_customers.csv

# Multi-stage transformation
zea load "sales/*.parquet" \
  | zea join customers.csv --on=customer_id --type=left \
  | zea filter "tier = 'Gold'" \
  | zea pivot --index=date --column=product --values=amount \
  | zea store gold_customer_sales.parquet
```

---

## Common Workflows

### Data Exploration

```bash
# Quick preview
zea load data.csv | zea describe

# Interactive exploration
zea view data.csv

# Check schema before loading large dataset
zea load "sales/" --schema-preview
```

### Filtering and Selecting

```bash
# Find high-value transactions
zea load sales.csv | zea filter "amount > 1000"

# Specific region data
zea load sales.csv | zea filter "region = 'west'" | zea select customer,amount
```

### Aggregation

```bash
# Sales by region
zea load sales.csv | zea group region --sum=amount

# Customer analysis
zea load sales.csv | zea group customer --sum=amount --count=1 --avg=amount
```

### Format Conversion

```bash
# CSV to Parquet
zea load data.csv | zea store data.parquet

# Parquet to JSON
zea load data.parquet | zea store data.json

# Multiple conversions
zea load input.csv | zea store temp.parquet
zea load temp.parquet | zea filter "amount > 100" | zea store output.json
```

### Joining Datasets

```bash
# Enrich sales with customer data
zea join sales.csv customers.csv --on=customer_id --type=left

# Find orphaned records
zea join orders.csv customers.csv --on=cust_id --type=left \
  | zea filter "customer_name IS NULL"
```

### Reshaping Data

```bash
# Long to wide
zea load sales_long.csv \
  | zea pivot --index=date --column=region --values=amount \
  | zea store sales_wide.parquet

# Wide to long
zea load sales_wide.csv \
  | zea unpivot --id=date --values=west,east,north,south --name=region --value=amount \
  | zea store sales_long.csv
```
