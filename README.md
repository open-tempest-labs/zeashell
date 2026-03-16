# ZeaShell

**DataFrame Shell - CSV to petabytes, one pipe at a time**

ZeaShell is a production-ready Go CLI for data processing with an embedded **ZeaFrame** DataFrame library. It's the first component of ZeaOS, a modern data shell inspired by PICK OS but built for modern data and table formats and processing with full Unix pipe semantics.

## Features

- **Pipeable Commands**: Full Unix pipe compatibility for data workflows
- **ZeaFrame Engine**: Embedded columnar DataFrame library
- **Multi-Format**: CSV, TSV, JSON, JSONL, and **Apache Parquet** support
- **Fast**: Single static binary, columnar storage, minimal dependencies
- **Expressive**: SQL-like filter expressions and aggregations
- **Production Ready**: Type inference, error handling, streaming I/O
- **Parquet Support**: Native Parquet read/write with Apache Arrow

## Installation

### Build from Source

```bash
git clone https://github.com/lmccay/zeashell
cd zeashell
go build -o zea ./cmd/zea
```

The `zea` binary is now ready to use!

### Install

```bash
go install github.com/lmccay/zeashell/cmd/zea@latest
```

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

### `zea load [file]`

Load CSV/TSV/JSON/JSONL/Parquet file and output to stdout. Format is auto-detected from file extension.

```bash
zea load sales.csv                    # Load CSV file
zea load data.tsv                     # Load TSV file
zea load data.json                    # Load JSON file (array of objects)
zea load events.jsonl                 # Load JSONL file (one object per line)
zea load sales.parquet                # Load Parquet file
cat sales.csv | zea load              # Load from stdin
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

Filter rows based on boolean expressions with support for nested field queries.

**Supported operators:**
- Comparison: `=`, `!=`, `>`, `>=`, `<`, `<=`
- Array membership: `CONTAINS`
- Logical: `AND`, `OR`

**Nested field support:**
- Array indexing: `orders[0] > 1000`
- Nested paths: `address.city = 'SF'`
- Array contains: `tags CONTAINS 'premium'`
- Wildcard patterns: `name CONTAINS '*.webshell.*'`

**Examples:**

```bash
# Simple filtering
zea load sales.csv | zea filter "amount > 100"
zea load sales.csv | zea filter "region = 'west'"
zea load sales.csv | zea filter "amount > 100 AND region = 'west'"
zea load sales.csv | zea filter "customer != '' AND amount >= 50"

# Nested field queries (JSON)
zea load data.json | zea filter "orders CONTAINS 1005"
zea load data.json | zea filter "orders[0] > 1004"
zea load data.json | zea filter "address.city = 'SF'"
zea load data.json | zea filter "address.state = 'CA' AND tags CONTAINS 'premium'"

# Wildcard pattern matching
zea load data.csv | zea filter "name CONTAINS '*.webshell.*'"
zea load data.csv | zea filter "service CONTAINS 'api.*'"
zea load data.json | zea filter "services CONTAINS '*.prod.?????'"
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

### `zea store [file]`

Store data to a file (or stdout). Format is auto-detected from file extension.

```bash
zea load sales.csv | zea filter "amount > 100" | zea store output.csv
zea load data.csv | zea select customer,total | zea store summary.tsv
zea load data.csv | zea filter "amount > 100" | zea store filtered.json
zea load events.csv | zea filter "status = 'active'" | zea store events.jsonl
zea load sales.csv | zea filter "amount > 1000" | zea store high_value.parquet
```

### `zea describe`

Show schema and preview of the data.

```bash
zea load sales.csv | zea describe
zea load data.csv | zea filter "amount > 100" | zea describe
```

## Format Conversion

ZeaShell supports seamless conversion between **all 5 formats**: CSV, TSV, JSON, JSONL, and Parquet.

```bash
# Any format to any format
zea load data.csv | zea store data.parquet      # CSV → Parquet
zea load data.parquet | zea store data.tsv      # Parquet → TSV
zea load data.tsv | zea store data.jsonl        # TSV → JSONL
zea load data.jsonl | zea store data.csv        # JSONL → CSV

# Full conversion chain
zea load input.csv | \
  zea store temp.tsv
zea load temp.tsv | \
  zea store temp.jsonl
zea load temp.jsonl | \
  zea store output.parquet
```

**All formats work interchangeably!** See [FORMAT_CONVERSION.md](FORMAT_CONVERSION.md) for complete guide.

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

ZeaShell supports a powerful expression language for filtering with full nested field support:

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

**Nested Field Access:**
- Array indexing: `field[0]`, `field[1]`, etc.
- Nested paths: `field.subfield.property`
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

# Nested field queries
orders CONTAINS 1005
orders[0] > 1000
address.city = 'SF'
tags CONTAINS 'premium' AND address.state = 'CA'

# Complex combinations
amount > 100 AND (region = 'west' OR region = 'east')
orders[0] > 1000 AND address.city != 'Oakland'

# Wildcard patterns
name CONTAINS '*.webshell.*'     # Match any with 'webshell' in middle
service CONTAINS 'api.*.prod'    # Match API services in prod
host CONTAINS 'server-??'        # Match server-01, server-02, etc.
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
