# ZeaShell

**DataFrame Shell - CSV to petabytes, one pipe at a time**

ZeaShell is a production-ready Go CLI for data processing with an embedded **ZeaFrame** DataFrame library. It's the first component of ZeaOS, a modern data shell inspired by PICK OS but built for modern data and table formats and processing with full Unix pipe semantics.

## Features

- **Pipeable Commands**: Full Unix pipe compatibility for data workflows
- **Interactive TUI**: Terminal-based data viewer with sort, filter, graph, and export
- **ZeaFrame Engine**: Embedded columnar DataFrame library
- **Multi-Format**: CSV, TSV, JSON, JSONL, XML and **Apache Parquet** support
- **Partitioned Data**: Glob patterns, directory loading, parallel multi-file operations
- **HTTP/HTTPS Support**: Load data directly from URLs
- **Relational Joins**: Inner, left, right, and full outer joins on multiple keys
- **Pivot/Unpivot**: Transform between long and wide formats
- **Fast**: Single static binary, columnar storage, minimal dependencies, parallel loading
- **Expressive**: SQL-like filter expressions and aggregations
- **Production Ready**: Type inference, error handling, streaming I/O, schema evolution
- **Parquet Support**: Native Parquet read/write with Apache Arrow
- **Path-Based Columns**: Nested JSON/XML structures flattened to dotted column names

## Path-Based Column Semantics

ZeaShell automatically flattens nested structures (JSON objects and XML hierarchies) into **path-based column names** using dot notation:

### JSON Flattening

```bash
# Input JSON with nested objects
$ cat data.json
[
  {"customer": "Alice", "address": {"city": "SF", "state": "CA"}},
  {"customer": "Bob", "address": {"city": "LA", "state": "CA"}}
]

# Flattened to path-based columns
$ zea load data.json
customer,address.city,address.state
Alice,SF,CA
Bob,LA,CA

# Filter using dotted paths
$ zea load data.json | zea filter "address.city = 'SF'"
```

### XML Flattening

```bash
# XML with hierarchical structure
$ cat topology.xml
<topology>
  <gateway>
    <provider>
      <role>authentication</role>
      <name>ShiroProvider</name>
    </provider>
  </gateway>
  <service>
    <role>WEBHDFS</role>
    <url>http://localhost:50070/webhdfs</url>
  </service>
</topology>

# Flattened to path-based columns
$ zea load topology.xml
gateway.provider.role,gateway.provider.name,service.role,service.url
authentication,ShiroProvider,WEBHDFS,http://localhost:50070/webhdfs

# Filter by path to show only services
$ zea load topology.xml | zea filter "service.role != ''"
```

**Benefits:**
- **No metadata pollution**: No format-specific columns like `_element`
- **Natural filtering**: Use dotted paths directly in expressions
- **Unified model**: Same semantics across JSON and XML
- **Path semantics**: Column names preserve hierarchical structure

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap open-tempest-labs/zeashell
brew install zeashell
```

### Using Go Install

```bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@latest
```

### Build from Source

```bash
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell
go build -o zea ./cmd/zea
```

The `zea` binary is now ready to use!

## Quick Start

```bash
# Load and explore data
zea load examples/sales.csv | zea describe

# Filter high-value sales
zea load examples/sales.csv | zea filter "amount > 100"

# Select specific columns
zea load examples/sales.csv | zea select customer,amount

# Group by region and sum amounts
zea load examples/sales.csv | zea group region --sum=amount

# Complete pipeline with Parquet output
zea load examples/sales.csv | \
  zea filter "region = 'west' AND amount > 100" | \
  zea select customer,amount | \
  zea group customer --sum=amount | \
  zea store summary.parquet

# Load Parquet files
zea load data.parquet | zea filter "amount > 1000"
```

## Commands

### `zea load [file|url|pattern|directory]`

Load files, glob patterns, directories, or URLs and output to stdout. Format is auto-detected from file extension.

**Supports:**
- Single files
- Glob patterns (`*.csv`, `sales/**/*.parquet`)
- Directories (recursive by default)
- Multiple files (comma-separated)
- HTTP/HTTPS URLs
- stdin

**Multi-file loading:**
- Files loaded in parallel for performance
- Schema inferred from first 3 files
- All files unioned into single table
- Missing columns filled with NULLs

**Flags:**
- `--max-files=N` - Limit number of files (default: unlimited)
- `--parallel=N` - Number of parallel workers (default: 8)
- `--format=fmt` - Filter by format (csv, parquet, json, etc.)
- `--schema-preview` - Show schema without loading data

```bash
# Single files
zea load sales.csv                    # Load CSV file
zea load data.parquet                 # Load Parquet file

# Glob patterns
zea load "*.csv"                      # All CSV files in current dir
zea load "sales/date=*.parquet"       # Partitioned Parquet files
zea load "sales/date=2026-03-*/*.csv" # Specific date partitions

# Directories (recursive by default)
zea load "sales/"                     # All supported files in sales/
zea load "sales/**/*.parquet"         # All Parquet files recursively

# Multiple files
zea load "file1.csv,file2.csv"        # Load multiple specific files
zea load "data/*.csv,archive/*.csv"   # Multiple glob patterns

# Schema preview
zea load "sales/" --schema-preview    # Show inferred schema

# Performance tuning
zea load "sales/" --parallel=16       # Use 16 workers
zea load "sales/" --max-files=100     # Limit to 100 files

# Remote URLs (HTTP/HTTPS)
zea load https://example.com/data.csv # Load CSV from URL
zea load https://data.gov/dataset.parquet  # Load Parquet from URL

# stdin
cat sales.csv | zea load              # Load from stdin

# Partitioned data workflows
zea load "sales/date=2026-03-*/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount \
  | zea store summary.parquet
```

### `zea select [columns]`

Select (project) specific columns.

```bash
zea load sales.csv | zea select customer,amount
zea load data.csv | zea select region,product,date
```

### `zea sort [columns]`

Sort rows by one or more columns with optional order specification.

**Column format**: `column[:asc|:desc]`
- `column` - Sort ascending (default)
- `column:asc` - Sort ascending (explicit)
- `column:desc` - Sort descending

**Multiple columns** are applied in order (stable sort).

**Examples:**

```bash
# Single column sorts
zea load sales.csv | zea sort amount              # Sort by amount ascending
zea load sales.csv | zea sort amount:desc         # Sort by amount descending

# Multi-column sorts
zea load sales.csv | zea sort region,amount       # Sort by region, then amount
zea load sales.csv | zea sort region,amount:desc  # Region asc, amount desc

# In pipelines
zea load sales.csv | \
  zea filter "amount > 100" | \
  zea sort region,amount:desc | \
  zea select region,customer,amount
```

### `zea filter [expression]`

Filter rows based on boolean expressions with support for path-based column names.

**Supported operators:**
- Comparison: `=`, `!=`, `>`, `>=`, `<`, `<=`
- Array membership: `CONTAINS`
- Logical: `AND`, `OR`

**Path-based columns:**
- Dotted paths from flattened JSON/XML: `address.city = 'SF'`
- Works naturally with nested structures: `gateway.provider.role = 'authentication'`
- Array indexing for legacy JSON strings: `orders[0] > 1000`
- Wildcard patterns: `service.role CONTAINS 'WEB*'`

**Examples:**

```bash
# Simple filtering
zea load sales.csv | zea filter "amount > 100"
zea load sales.csv | zea filter "region = 'west'"
zea load sales.csv | zea filter "amount > 100 AND region = 'west'"
zea load sales.csv | zea filter "customer != '' AND amount >= 50"

# Path-based filtering (JSON/XML)
zea load data.json | zea filter "address.city = 'SF'"
zea load data.json | zea filter "address.state = 'CA' AND tags CONTAINS 'premium'"
zea load topology.xml | zea filter "service.role = 'WEBHDFS'"
zea load topology.xml | zea filter "gateway.provider.role = 'authentication'"

# Wildcard pattern matching
zea load data.csv | zea filter "name CONTAINS '*.webshell.*'"
zea load topology.xml | zea filter "service.role CONTAINS 'WEB*'"
zea load data.json | zea filter "services CONTAINS '*.prod.?????'"

# Array operations (for array-valued columns)
zea load data.json | zea filter "orders CONTAINS 1005"
```

### `zea group [columns] [--agg=col]`

Group by columns and perform aggregations.

**Supported aggregations:**
- `--sum=column`: Sum of values
- `--avg=column`: Average of values
- `--min=column`: Minimum value
- `--max=column`: Maximum value
- `--count=column`: Count of rows (use `1` for row count)

**Examples:**

```bash
# Total sales by region
zea load sales.csv | zea group region --sum=amount

# Product statistics
zea load sales.csv | zea group product --sum=amount --count=1

# Multiple grouping columns
zea load sales.csv | zea group region,product --sum=amount --avg=amount

# Multiple aggregations
zea load sales.csv | zea group customer --sum=amount --count=1 --avg=amount
```

### `zea join [left-source] [right-source]`

Join two datasets on one or more key columns.

**Join types:**
- `inner` - Only rows with matches in both datasets (default)
- `left` - All left rows, NULLs for unmatched right
- `right` - All right rows, NULLs for unmatched left
- `full` - All rows from both, NULLs where no match

**Flags:**
- `--on=column[,column2,...]` - Join key column(s) (required)
- `--type=inner|left|right|full` - Join type (default: inner)

**Column name collisions** are resolved by adding `_right` suffix to right-side columns.

**Examples:**

```bash
# Inner join on single key
zea join customers.csv orders.csv --on=cust_id

# Left join on multiple keys
zea join customers.csv orders.csv --on=id,date --type=left

# Join stdin with file
zea load customers.csv | zea join orders.csv --on=cust_id

# Join and filter
zea join customers.csv orders.csv --on=cust_id --type=left \
  | zea filter "order_id IS NULL"

# Join from remote URLs
zea join https://data.example.com/customers.csv \
  https://data.example.com/orders.csv --on=cust_id

# Complex pipeline with join
zea load sales.csv \
  | zea filter "amount > 100" \
  | zea join products.csv --on=product_id \
  | zea select customer,product.name,amount \
  | zea store enriched.parquet
```

### `zea pivot`

Transform long format data to wide format.

**Flags:**
- `--index=column[,column2,...]` - Index column(s) to group by (required)
- `--column=column` - Column whose values become new column names (required)
- `--values=column` - Column whose values populate the new columns (required)

**Example transformation:**
```
Input (long):                Output (wide):
date,region,amount          date,west,east
2026-01-01,west,100         2026-01-01,100,50
2026-01-01,east,50          2026-01-02,70,
2026-01-02,west,70
```

**Examples:**

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

### `zea unpivot`

Transform wide format data to long format.

**Flags:**
- `--id=column[,column2,...]` - ID column(s) to preserve (optional)
- `--values=column[,column2,...]` - Columns to unpivot into rows (required)
- `--name=column` - Name for column containing original column names (default: variable)
- `--value=column` - Name for column containing values (default: value)

**Example transformation:**
```
Input (wide):                 Output (long):
date,west,east               date,region,amount
2026-01-01,100,50            2026-01-01,west,100
2026-01-02,70,               2026-01-01,east,50
                             2026-01-02,west,70
```

**Examples:**

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

### `zea store [file]`

Store data to a file (or stdout). Format is auto-detected from file extension.

```bash
zea load sales.csv | zea filter "amount > 100" | zea store output.csv
zea load data.csv | zea select customer,total | zea store summary.tsv
zea load data.csv | zea filter "amount > 100" | zea store filtered.json
zea load events.csv | zea filter "status = 'active'" | zea store events.jsonl
zea load topology.xml | zea filter "service.role != ''" | zea store services.xml
zea load sales.csv | zea filter "amount > 1000" | zea store high_value.parquet
```

### `zea describe`

Show schema and preview of the data.

```bash
zea load sales.csv | zea describe
zea load data.csv | zea filter "amount > 100" | zea describe
```

### `zea view [source]`

Interactive terminal UI for data exploration with keyboard-driven operations.

**Features:**
- Scrollable table view with syntax highlighting
- Sort by column (ascending/descending)
- Filter with expressions
- Graph/chart view (histograms for numeric, bar charts for categorical)
- Export filtered/sorted view to CSV
- Full keyboard navigation

**Keyboard Shortcuts:**

Navigation:
- `↑↓←→` - Move cursor
- `PgUp/PgDn` - Page up/down
- `Home/End` - First/last row

Operations:
- `s` - Sort by current column (cycle: none → asc → desc)
- `f` - Filter with expression dialog
- `g` - Show graph/chart for column
- `e` - Export current view to file
- `r` - Reset filters and sorts
- `?` - Show help overlay
- `q` - Quit

**Examples:**

```bash
# View single file
zea view sales.csv

# View partitioned data
zea view "sales/**/*.parquet"

# View from URL
zea view "https://example.com/data.csv"

# View pipeline output
zea load sales.csv | zea filter "amount > 100" | zea view -

# View with glob pattern
zea view "testdata/sales-partitioned/date=*/*.csv"
```

**Use cases:**
- Quick data exploration without writing code
- Verify filter/sort operations interactively
- Generate graphs for numeric distributions
- Export subsets after interactive filtering
- Explore partitioned data lakes

## Format Conversion

ZeaShell supports seamless conversion between **all 6 formats**: CSV, TSV, JSON, JSONL, XML, and Parquet.

```bash
# Any format to any format
zea load data.csv | zea store data.parquet      # CSV → Parquet
zea load data.parquet | zea store data.tsv      # Parquet → TSV
zea load data.tsv | zea store data.jsonl        # TSV → JSONL
zea load data.jsonl | zea store data.csv        # JSONL → CSV
zea load topology.xml | zea store data.csv      # XML → CSV (flattened)
zea load data.json | zea store topology.xml     # JSON → XML

# Full conversion chain
zea load input.csv | \
  zea store temp.tsv
zea load temp.tsv | \
  zea store temp.jsonl
zea load temp.jsonl | \
  zea store output.parquet
```

**All formats work interchangeably!** Nested JSON and XML structures are automatically flattened to path-based columns.

## HTTP/HTTPS Support

ZeaShell can load data directly from HTTP and HTTPS URLs, making it easy to work with remote datasets:

```bash
# Load and analyze remote CSV data
zea load "https://raw.githubusercontent.com/datasets/gdp/master/data/gdp.csv" | \
  zea filter "Year > 2015" | \
  zea select "Country Name,Year,Value" | \
  head -10

# Load JSON from API and filter
zea load "https://api.example.com/users.json" | \
  zea filter "address.city = 'SF'" | \
  zea group address.state --count=1

# Convert remote data to local format
zea load "https://data.gov/dataset.csv" | \
  zea filter "year = 2023" | \
  zea store local_2023.parquet

# Chain remote sources
zea load "https://example.com/sales.csv" | \
  zea filter "amount > 1000" | \
  zea group region --sum=amount
```

**Supported URL features:**
- Automatic format detection from URL file extension
- HTTP and HTTPS protocols
- 30-second timeout for remote requests
- Streaming for CSV, TSV, JSON, JSONL, XML formats
- Temporary file download for Parquet (requires seekable access)

## Partitioned Data & Globbing

ZeaShell excels at loading **partitioned data** with glob patterns and directory traversal - perfect for data lake workflows and Hive-style partitioning.

### Partitioned Data Loading

```bash
# Directory structure:
# sales/
# ├── date=2026-03-01/sales.parquet
# ├── date=2026-03-02/sales.parquet
# └── date=2026-03-03/sales.parquet

# Load entire directory (all partitions)
zea load "sales/"

# Load specific date partitions
zea load "sales/date=2026-03-01/*.parquet"

# Load date range with glob
zea load "sales/date=2026-03-*/*.parquet"

# Nested partitions
zea load "sales/year=2026/month=03/**/*.parquet"
```

### Glob Patterns

```bash
# Simple wildcards
zea load "*.csv"                      # All CSV in current directory
zea load "data/*.parquet"             # All Parquet in data/

# Recursive glob
zea load "sales/**/*.csv"             # All CSV files recursively

# Multiple patterns
zea load "jan/*.csv,feb/*.csv"        # Comma-separated patterns

# Complex patterns
zea load "sales/date=2026-*/region=*/*.parquet"
```

### Performance Features

**Parallel Loading:**
```bash
# 8 workers by default
zea load "sales/"

# Tune parallelism
zea load "sales/" --parallel=16

# Limit files for safety
zea load "sales/" --max-files=1000
```

**Schema Evolution:**
```bash
# Files with different schemas are automatically unioned
# Missing columns are filled with NULLs

# sales/2026-01.csv: id,amount
# sales/2026-02.csv: id,amount,region

zea load "sales/*.csv"
# Result has all columns: id, amount, region
# Rows from 2026-01 have NULL for region
```

### Real-World Workflows

**Analyze Partitioned Sales Data:**
```bash
zea load "sales/date=2026-03-*/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount --count=1 \
  | zea store sales_summary.parquet
```

**Join Partitioned Data with Dimension Tables:**
```bash
zea load "transactions/date=*/*.parquet" \
  | zea join "dimensions/customers.csv" --on=customer_id \
  | zea filter "tier = 'Gold'" \
  | zea group product --sum=amount
```

**Time-Series Pivot:**
```bash
zea load "metrics/date=2026-*/*.csv" \
  | zea pivot --index=metric --column=date --values=value \
  | zea store metrics_wide.parquet
```

**Schema Preview (No Loading):**
```bash
# Quickly check schema before loading large datasets
zea load "sales/" --schema-preview

# Output:
# Found 156 files
#
# Inferred schema:
#   customer: string
#   region: string
#   amount: int64
#   date: string
#   product: string
```

### Cloud Storage via Volumez Mounts

ZeaShell works seamlessly with cloud storage through filesystem mounts (Volumez, FUSE, etc.):

```bash
# S3 bucket mounted at /mnt/s3-data via Volumez
zea load "/mnt/s3-data/sales/date=2026-*/*.parquet"

# Azure Blob mounted at /mnt/azure
zea load "/mnt/azure/warehouse/**/*.csv"

# Local, S3, GCS - all the same to ZeaShell
zea load "/mnt/*/sales/*.parquet" \
  | zea filter "amount > 1000" \
  | zea store /mnt/local/summary.parquet
```

**Benefits:**
- Cloud-agnostic (S3, Azure, GCS, MinIO)
- Standard Unix permissions and tooling
- Volumez handles optimization (caching, prefetching)
- No cloud SDK dependencies in ZeaShell
- Use standard Unix tools: `ls`, `find`, `du` work on cloud data

**Example: Petabyte-scale analysis**
```bash
# S3 data lake mounted via Volumez at /mnt/datalake
# 10TB of partitioned Parquet across 1000+ files

zea load "/mnt/datalake/events/year=2026/month=03/**/*.parquet" \
  | zea filter "user_tier = 'premium' AND revenue > 100" \
  | zea group country --sum=revenue --count=1 \
  | zea pivot --index=country --column=product --values=sum_revenue \
  | zea store /mnt/local/premium_revenue_2026_03.parquet

# ZeaShell handles:
# - Parallel loading (1000+ files)
# - Schema evolution (files may have different columns)
# - Memory-efficient streaming

# Volumez handles:
# - S3 authentication
# - Intelligent caching
# - Parallel S3 reads
# - Network optimization
```

## Examples

### 1. Find Top Customers by Sales

```bash
zea load examples/sales.csv | \
  zea group customer --sum=amount | \
  zea filter "amount_sum > 1000"
```

### 2. Regional Analysis

```bash
zea load examples/sales.csv | \
  zea group region --sum=amount --count=1 --avg=amount
```

### 3. Product Performance

```bash
zea load examples/sales.csv | \
  zea filter "amount > 50" | \
  zea group product --sum=amount --count=1
```

### 4. West Region Laptop Sales

```bash
zea load examples/sales.csv | \
  zea filter "region = 'west' AND product = 'laptop'" | \
  zea select customer,amount,date
```

### 5. Export Filtered Data

```bash
zea load examples/sales.csv | \
  zea filter "amount > 100" | \
  zea select customer,region,product,amount | \
  zea store high_value_sales.csv
```

### 6. Multi-Column Grouping

```bash
zea load examples/sales.csv | \
  zea group region,product --sum=amount --count=1 | \
  zea filter "amount_sum > 500"
```

### 7. Customer Segmentation

```bash
zea load examples/sales.csv | \
  zea filter "region = 'west' OR region = 'east'" | \
  zea group customer,region --sum=amount --count=1
```

### 8. Date-Based Filtering

```bash
zea load examples/sales.csv | \
  zea filter "date >= '2026-01-20'" | \
  zea group customer --sum=amount
```

### 9. Complex Pipeline

```bash
zea load examples/sales.csv | \
  zea filter "amount > 100 AND region = 'west'" | \
  zea select customer,product,amount | \
  zea group customer --sum=amount --count=1 | \
  zea filter "amount_sum > 1000" | \
  zea store west_vip_customers.csv
```

### 10. Unix Integration

```bash
# Combine with grep
cat examples/sales.csv | zea load | grep "Alice" | zea group customer --sum=amount

# Use with wc
zea load examples/sales.csv | zea filter "amount > 1000" | wc -l

# Chain with other tools
zea load examples/sales.csv | \
  zea filter "region = 'west'" | \
  cut -d, -f1,4 | \
  sort -t, -k2 -nr | \
  head -n 5
```

## Parquet Support (Phase 2)

ZeaShell now includes native Apache Parquet support! See [PARQUET.md](PARQUET.md) for detailed documentation.

### Quick Parquet Examples

```bash
# Convert CSV to Parquet
zea load sales.csv | zea store sales.parquet

# Query Parquet files
zea load sales.parquet | zea filter "amount > 1000"

# Cross-format pipeline
zea load data.csv | zea filter "region = 'west'" | zea store west.parquet
zea load west.parquet | zea group product --sum=amount | zea store summary.csv

# Run the Parquet demo
./examples/parquet-pipeline.sh
```

**Benefits**:
- **5-10x smaller** file sizes vs CSV
- **2-3x faster** processing with columnar storage
- **Full compatibility** with all ZeaShell commands
- **Format auto-detection** - just use `.parquet` extension

## ZeaFrame Library

ZeaShell embeds the **ZeaFrame** DataFrame library, which can be used programmatically in Go:

```go
import "github.com/open-tempest-labs/zeashell/internal/zeaframe"

// Load from CSV
file, _ := os.Open("sales.csv")
zf, _ := zeaframe.FromCSV(file)

// Select columns
zf = zf.Select("customer", "amount")

// Filter rows
expr, _ := zeaframe.ParseExpression("amount > 100")
zf = zf.Filter(expr)

// Group and aggregate
result, _ := zf.GroupBy("customer").Agg(map[string]string{
    "amount": "sum",
})

// Write output
result.WriteCSV(os.Stdout)
```

### Column Types

ZeaFrame automatically infers column types:
- `string`: Text data
- `int64`: Integer numbers
- `float64`: Decimal numbers
- `bool`: Boolean values (true/false)

## Expression Language

ZeaShell supports a powerful expression language for filtering with **path-based column names**:

**Comparison Operators:**
- `=` - Equal
- `!=` - Not equal
- `>` - Greater than
- `>=` - Greater than or equal
- `<` - Less than
- `<=` - Less than or equal

**Array Operators:**
- `CONTAINS` - Array membership and wildcard matching
  - Exact match: `tags CONTAINS 'premium'`
  - Wildcard: `name CONTAINS '*.webshell.*'`
  - Supports `*` (any characters) and `?` (single character)

**Logical Operators:**
- `AND` - Logical AND
- `OR` - Logical OR

**Path-Based Column Names:**
- Dotted paths from flattened JSON/XML: `address.city`, `service.role`, `gateway.provider.name`
- Use directly in expressions: `address.city = 'SF'`
- Array indexing for legacy JSON: `orders[0] > 1000`
- Works with all comparison operators

**Value Types:**
- Numbers: `100`, `99.99`
- Strings: `'west'`, `"laptop"`
- Booleans: `true`, `false`

**Examples:**
```
# Simple comparisons
amount > 100
region = 'west'
customer != ''
amount >= 50 AND region = 'CA'

# Path-based column filtering (JSON/XML)
address.city = 'SF'
address.state = 'CA' AND tags CONTAINS 'premium'
service.role = 'WEBHDFS'
gateway.provider.role = 'authentication'

# Complex combinations
amount > 100 AND (region = 'west' OR region = 'east')
service.role CONTAINS 'WEB*' AND service.url CONTAINS 'localhost'

# Wildcard patterns
name CONTAINS '*.webshell.*'     # Match any with 'webshell' in middle
service.role CONTAINS 'WEB*'     # Match WEBHDFS, WEBHCAT, WEBHBASE
host CONTAINS 'server-??'        # Match server-01, server-02, etc.

# Array operations (for array-valued columns)
orders CONTAINS 1005
tags CONTAINS 'premium'
```

## Performance

ZeaShell is designed for performance:
- **Streaming I/O**: Memory-efficient processing
- **Columnar Storage**: Fast aggregations and filtering
- **Static Binary**: No runtime dependencies
- **Optimized Parsing**: Efficient CSV/TSV/JSONL parsing

Expected performance:
- Process 1GB CSV files in <30 seconds
- Handle millions of rows
- Low memory footprint with streaming operations

## Architecture

```
zeashell/
├── cmd/zea/              # CLI entry point
├── internal/
│   ├── zeaframe/         # DataFrame library
│   │   ├── dataframe.go  # Core ZeaFrame engine
│   │   ├── parser.go     # Expression parser
│   │   ├── io.go         # CSV/TSV/JSONL I/O
│   │   ├── parquet.go    # Parquet I/O with Arrow
│   │   └── autodetect.go # Format auto-detection
│   └── cli/              # Command implementations
│       ├── root.go       # Root command
│       ├── load.go       # Load command
│       ├── select.go     # Select command
│       ├── filter.go     # Filter command
│       ├── sort.go       # Sort command
│       ├── group.go      # Group command
│       ├── store.go      # Store command
│       └── describe.go   # Describe command
└── examples/             # Sample data and scripts
```

## Dependencies

ZeaShell uses minimal dependencies:
- Go standard library
- `github.com/spf13/cobra` - CLI framework
- `github.com/apache/arrow/go/v18` - Apache Arrow (for Parquet support)

## Roadmap

**Phase 1 (MVP - Completed):**
- ✅ Core DataFrame operations (select, filter, group)
- ✅ CSV/TSV/JSONL support
- ✅ Expression parser
- ✅ Unix pipe compatibility
- ✅ Aggregations (sum, avg, min, max, count)

**Phase 2 (Completed):**
- ✅ **Parquet read/write** with Apache Arrow
- ✅ **Format auto-detection**
- ✅ **Cross-format pipelines**
- ✅ Columnar storage optimization

**Phase 3 (Planned):**
- Join operations between datasets
- Parquet glob patterns (`sales/*.parquet`)
- S3 path support (`s3://bucket/data.parquet`)
- Predicate pushdown for Parquet
- Interactive REPL mode
- Config file management

**Phase 4 (Future):**
- Iceberg catalog integration
- Window functions (rank, lag, lead)
- Multi-file partitioned datasets
- Parallel processing
- Query optimization
- ZeaCatalog integration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache License 2.0 - see LICENSE file for details.

## Inspiration

ZeaShell is inspired by:
- **PICK OS**: Multi-valued databases and interactive shells
- **Unix Philosophy**: Do one thing well, composability
- **Modern DataFrames**: Pandas, Polars, DataFusion
- **DuckDB**: Fast analytical queries on CSV files

## Authors

Created by the ZeaOS team.

## Links

- GitHub: https://github.com/lmccay/zeashell
- Issues: https://github.com/lmccay/zeashell/issues

---

**Start processing data with ZeaShell today!** 🌊🐚
