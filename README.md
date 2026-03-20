# ZeaShell

**DataFrame Shell - CSV to petabytes, one pipe at a time**

ZeaShell is a production-ready Go CLI for data processing with an embedded **ZeaFrame** DataFrame library. Process CSV, JSON, XML, and Parquet files with Unix pipe semantics - from quick data exploration to petabyte-scale data lake workflows.

## Features

- 🔄 **Pipeable Commands** - Full Unix pipe compatibility for data workflows
- 📊 **Interactive TUI** - Terminal viewer with sort, filter, graph, and export
- 🚀 **Multi-Format** - CSV, TSV, JSON, JSONL, XML, and Apache Parquet support
- 🗂️ **Partitioned Data** - Glob patterns, directory loading, parallel multi-file operations
- 🌐 **HTTP/HTTPS** - Load data directly from URLs
- 🔗 **Relational Joins** - Inner, left, right, and full outer joins
- ↔️ **Pivot/Unpivot** - Transform between long and wide formats
- 🔌 **Plugin System** - Extend with custom commands using simple executable scripts
- ⚡ **Fast** - Single static binary, columnar storage, parallel loading
- 🎯 **Expressive** - SQL-like filter expressions and aggregations
- 🏗️ **Production Ready** - Type inference, error handling, streaming I/O, schema evolution

<img width="2084" height="1976" alt="image" src="https://github.com/user-attachments/assets/c7d690e3-410c-4788-9ad1-cdb80e3956b8" />


## Quick Start

### Installation

**Homebrew (macOS/Linux):**
```bash
brew tap open-tempest-labs/zeashell
brew install zeashell
```

**Go Install:**
```bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@latest
```

**From Source:**
```bash
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell
go build -o zea ./cmd/zea
```

### Basic Usage

```bash
# Interactive exploration
zea view sales.csv

# Filter and select
zea load sales.csv | zea filter "amount > 100" | zea select customer,amount

# Group and aggregate
zea load sales.csv | zea group region --sum=amount --count=1

# Complete pipeline
zea load sales.csv \
  | zea filter "region = 'west' AND amount > 100" \
  | zea group customer --sum=amount \
  | zea store summary.parquet

# Load partitioned data
zea load "sales/date=2026-03-*/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount
```

## Commands

| Command | Description | Documentation |
|---------|-------------|---------------|
| `zea load` | Load data from files, URLs, patterns, or directories | [Commands](docs/COMMANDS.md#zea-load) |
| `zea view` | Interactive terminal UI viewer | [Viewer](docs/VIEWER.md) |
| `zea filter` | Filter rows with expressions | [Expressions](docs/EXPRESSIONS.md) |
| `zea select` | Select specific columns | [Commands](docs/COMMANDS.md#zea-select) |
| `zea sort` | Sort by one or more columns | [Commands](docs/COMMANDS.md#zea-sort) |
| `zea group` | Group by and aggregate | [Commands](docs/COMMANDS.md#zea-group) |
| `zea join` | Join two DataFrames | [Commands](docs/COMMANDS.md#zea-join) |
| `zea pivot` | Transform long to wide format | [Commands](docs/COMMANDS.md#zea-pivot) |
| `zea unpivot` | Transform wide to long format | [Commands](docs/COMMANDS.md#zea-unpivot) |
| `zea describe` | Show schema and preview | [Commands](docs/COMMANDS.md#zea-describe) |
| `zea store` | Write data to file | [Commands](docs/COMMANDS.md#zea-store) |
| `zea run` | Run custom plugin commands | [Plugins](docs/PLUGINS.md) |

## Key Features

### Interactive Viewer

Launch an interactive terminal UI to explore data visually:

```bash
zea view sales.csv
```

- **Navigate**: Arrow keys, PgUp/PgDn
- **Sort**: Press `s` on any column
- **Filter**: Press `f` to apply expressions
- **Graph**: Press `g` for charts (histograms, bar charts)
- **Export**: Press `e` to save filtered data
- **Help**: Press `?` for all shortcuts

[📖 Full Viewer Documentation](docs/VIEWER.md)

### Filter Expressions

Powerful SQL-like expressions with path-based column names:

```bash
# Simple comparisons
zea filter "amount > 1000"
zea filter "region = 'West' AND amount > 100"

# Path-based columns (from JSON/XML)
zea filter "address.city = 'SF'"
zea filter "service.role = 'WEBHDFS'"

# Pattern matching
zea filter "name CONTAINS '*.webshell.*'"
```

[📖 Expression Language Reference](docs/EXPRESSIONS.md)

### Partitioned Data & Globbing

Load multiple files with glob patterns - perfect for data lakes:

```bash
# Glob patterns
zea load "*.csv"
zea load "sales/**/*.parquet"

# Partitioned data (Hive-style)
zea load "sales/date=2026-03-*/*.parquet"
zea load "warehouse/year=2026/month=03/**/*.parquet"

# Performance options
zea load "sales/*.csv" --parallel=16 --max-files=1000

# Schema preview (no data loading)
zea load "sales/" --schema-preview
```

**Features:**
- Parallel loading (configurable workers)
- Schema evolution (missing columns → NULLs)
- Type promotion (int → float → string)
- Works with cloud storage via Volumez mounts

[📖 Partitioned Data Guide](docs/PARTITIONED_DATA.md)

### Multi-Format Support

Seamless conversion between all formats:

```bash
# Format auto-detection by extension
zea load data.csv | zea store data.parquet      # CSV → Parquet
zea load data.parquet | zea store data.json     # Parquet → JSON
zea load topology.xml | zea store data.csv      # XML → CSV (flattened)

# Load from URLs
zea load "https://example.com/data.csv" | zea filter "amount > 100"
```

**Supported formats:**
- CSV, TSV
- JSON, JSONL
- XML
- Apache Parquet

[📖 Parquet Documentation](PARQUET.md)

### Path-Based Columns

Nested JSON/XML structures are automatically flattened to dotted column names:

```json
{
  "customer": "Alice",
  "address": {"city": "SF", "state": "CA"}
}
```

Becomes columns: `customer`, `address.city`, `address.state`

```bash
# Filter using dotted paths
zea load data.json | zea filter "address.city = 'SF'"
```

**Benefits:**
- No metadata pollution
- Natural filtering with dotted paths
- Unified model across JSON and XML
- Path semantics preserved

### Plugin System

Extend ZeaShell with custom commands using simple executable scripts:

```bash
# Create a plugin directory
mkdir -p ~/.zea/plugins

# Create an executable plugin script
cat > ~/.zea/plugins/sales << 'EOF'
#!/bin/bash
# @desc Process sales CSV files with standard transforms

zea load "$1" | zea filter "amount > 1000" | zea view
EOF

chmod +x ~/.zea/plugins/sales

# Run your plugin
zea run sales data.csv
```

**Features:**
- Drop executable scripts into `~/.zea/plugins/`
- Automatic command registration as `zea run <plugin>`
- Full stdin/stdout/stderr passthrough
- Metadata directives for help text (`@desc`, `@name`, `@args`)
- Works with any scripting language (Bash, Python, Ruby, etc.)
- Seamless pipeline integration
- Auto-completion support

[📖 Plugin System Documentation](docs/PLUGINS.md)

## Example Workflows

### Data Exploration

```bash
# Quick preview
zea load data.csv | zea describe

# Interactive exploration
zea view data.csv

# Check schema of large dataset
zea load "sales/" --schema-preview
```

### Analysis Pipeline

```bash
# High-value West region customers
zea load sales.csv \
  | zea filter "amount > 1000 AND region = 'west'" \
  | zea group customer --sum=amount --count=1 \
  | zea filter "amount_sum > 5000" \
  | zea sort amount_sum:desc \
  | zea store top_customers.csv
```

### Join and Pivot

```bash
# Enrich sales with customer data, then pivot by region
zea load sales.csv \
  | zea join customers.csv --on=customer_id --type=left \
  | zea filter "tier = 'Gold'" \
  | zea pivot --index=date --column=region --values=amount \
  | zea store gold_sales_by_region.parquet
```

### Partitioned Data Lake

```bash
# Analyze S3 data mounted via Volumez
zea load "/mnt/datalake/events/year=2026/month=03/**/*.parquet" \
  | zea filter "user_tier = 'premium' AND revenue > 100" \
  | zea group country --sum=revenue --count=1 \
  | zea store /mnt/local/premium_revenue.parquet
```

## Documentation

### Getting Started
- [README](README.md) - This file
- [Installation](#installation) - Get up and running
- [Quick Start](#quick-start) - Basic usage examples

### Core Concepts
- [Commands Reference](docs/COMMANDS.md) - Complete command documentation
- [Expression Language](docs/EXPRESSIONS.md) - Filter syntax and operators
- [Interactive Viewer](docs/VIEWER.md) - TUI viewer guide
- [Partitioned Data](docs/PARTITIONED_DATA.md) - Glob patterns and multi-file loading
- [Plugin System](docs/PLUGINS.md) - Extend ZeaShell with custom commands

### Format Support
- [Parquet Support](PARQUET.md) - Apache Parquet documentation
- [Filter Syntax](examples/FILTER_SYNTAX.md) - Filter examples and troubleshooting

### Examples
- [Example Data](examples/) - Sample datasets
- [Glob Demo](examples/glob-demo.sh) - Multi-file loading demos
- [TUI Mockup](examples/TUI_MOCKUP.md) - Visual mockups of viewer

## Architecture

```
zeashell/
├── cmd/zea/              # CLI entry point
├── internal/
│   ├── zeaframe/         # DataFrame library
│   │   ├── dataframe.go  # Core engine
│   │   ├── parser.go     # Expression parser
│   │   ├── io.go         # CSV/TSV/JSON I/O
│   │   ├── parquet.go    # Parquet I/O
│   │   ├── join.go       # Join operations
│   │   ├── pivot.go      # Pivot/unpivot
│   │   ├── glob.go       # Pattern matching
│   │   └── union.go      # Multi-file loading
│   ├── cli/              # Commands
│   └── tui/              # Interactive viewer
├── docs/                 # Documentation
└── examples/             # Sample data and demos
```

## Dependencies

ZeaShell uses minimal, well-maintained dependencies:

**Core:**
- Go standard library
- `github.com/spf13/cobra` - CLI framework

**Data Processing:**
- `github.com/apache/arrow/go/v18` - Apache Arrow (for Parquet support)

**Interactive TUI:**
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/rivo/tview` - Terminal UI framework

All dependencies are vendored for reproducible builds.

## Performance

ZeaShell is designed for performance:
- **Streaming I/O** - Memory-efficient processing
- **Columnar Storage** - Fast aggregations and filtering
- **Parallel Loading** - Multi-threaded file loading
- **Static Binary** - No runtime dependencies

Expected performance:
- Process 1GB CSV files in <30 seconds
- Handle millions of rows
- Parquet 2-3x faster than CSV
- Low memory footprint with streaming

## Cloud Storage

ZeaShell works with cloud storage through filesystem mounts (Volumez, FUSE, etc.):

```bash
# S3 bucket mounted at /mnt/s3-data via Volumez
zea load "/mnt/s3-data/sales/date=2026-*/*.parquet"

# Local, S3, GCS - all the same to ZeaShell
zea load "/mnt/*/sales/*.parquet" | zea filter "amount > 1000"
```

**Benefits:**
- Cloud-agnostic (S3, Azure, GCS, MinIO)
- Standard Unix tools work on cloud data
- Volumez handles caching and optimization
- No cloud SDK dependencies

## ZeaFrame Library

ZeaShell embeds the **ZeaFrame** DataFrame library for programmatic use:

```go
import "github.com/open-tempest-labs/zeashell/internal/zeaframe"

// Load from CSV
file, _ := os.Open("sales.csv")
zf, _ := zeaframe.FromCSV(file)

// Filter and select
expr, _ := zeaframe.ParseExpression("amount > 100")
zf, _ = zf.Filter(expr)
zf = zf.Select("customer", "amount")

// Group and aggregate
result, _ := zf.GroupBy("customer").Agg(map[string]string{
    "amount": "sum",
})

// Write output
result.WriteCSV(os.Stdout)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## Inspiration

ZeaShell is inspired by:
- **Multi-valued databases** - Hierarchical data models that match business reality
- **Interactive data shells** - Integrated environments for exploration and analysis
- **Unix Philosophy** - Do one thing well, composability through pipes
- **Modern DataFrames** - Pandas, Polars, DataFusion
- **DuckDB** - Fast analytical queries on data files

See [PICK Reimagined](docs/PICK_REIMAGINED.md) for the historical context and how classic database concepts inspire modern data processing

## Links

- **GitHub**: https://github.com/open-tempest-labs/zeashell
- **Issues**: https://github.com/open-tempest-labs/zeashell/issues
- **Homebrew Tap**: https://github.com/open-tempest-labs/homebrew-zeashell

---

**Start processing data with ZeaShell today!** 🌊🐚
