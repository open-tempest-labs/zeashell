# Partitioned Data & Glob Patterns

ZeaShell excels at loading **partitioned data** with glob patterns and directory traversal - perfect for data lake workflows and Hive-style partitioning.

## Quick Start

```bash
# Load all files in directory
zea load "sales/"

# Load with glob pattern
zea load "sales/*.csv"

# Load partitioned data
zea load "sales/date=2026-03-*/*.parquet"
```

## Glob Patterns

### Basic Wildcards

```bash
# All CSV files in current directory
zea load "*.csv"

# All Parquet files in sales directory
zea load "sales/*.parquet"

# Specific pattern
zea load "data-*.json"
```

### Recursive Glob

```bash
# All CSV files recursively
zea load "sales/**/*.csv"

# All Parquet files in any subdirectory
zea load "warehouse/**/*.parquet"
```

### Multiple Patterns

```bash
# Comma-separated patterns
zea load "jan/*.csv,feb/*.csv,mar/*.csv"

# Multiple directories
zea load "east/*.csv,west/*.csv"
```

### Complex Patterns

```bash
# Date partitions
zea load "sales/date=2026-*/*.parquet"

# Multi-level partitions
zea load "sales/year=2026/month=03/**/*.parquet"

# Specific regions and dates
zea load "sales/date=2026-*/region=*/*.parquet"
```

## Partitioned Data

### Hive-Style Partitioning

ZeaShell supports Hive-style partitioned data:

```
sales/
├── date=2026-03-01/
│   ├── sales.parquet
│   └── metadata.json
├── date=2026-03-02/
│   ├── sales.parquet
│   └── metadata.json
└── date=2026-03-03/
    ├── sales.parquet
    └── metadata.json
```

**Load all partitions:**
```bash
zea load "sales/"
```

**Load specific partitions:**
```bash
# Single date
zea load "sales/date=2026-03-01/*.parquet"

# Date range
zea load "sales/date=2026-03-*/*.parquet"
```

### Multi-Level Partitions

```
warehouse/
└── year=2026/
    └── month=03/
        ├── day=01/
        │   └── data.parquet
        ├── day=02/
        │   └── data.parquet
        └── day=03/
            └── data.parquet
```

**Load patterns:**
```bash
# Entire year
zea load "warehouse/year=2026/**/*.parquet"

# Specific month
zea load "warehouse/year=2026/month=03/**/*.parquet"

# Specific day
zea load "warehouse/year=2026/month=03/day=01/*.parquet"

# Multiple months
zea load "warehouse/year=2026/month=0[1-3]/**/*.parquet"
```

### Regional Partitioning

```
sales/
├── region=west/
│   ├── 2026-03-01.csv
│   └── 2026-03-02.csv
├── region=east/
│   ├── 2026-03-01.csv
│   └── 2026-03-02.csv
└── region=south/
    ├── 2026-03-01.csv
    └── 2026-03-02.csv
```

**Load patterns:**
```bash
# All regions
zea load "sales/**/*.csv"

# Specific region
zea load "sales/region=west/*.csv"

# Multiple regions
zea load "sales/region=west/*.csv,sales/region=east/*.csv"
```

## Performance Features

### Parallel Loading

Files are loaded in parallel for better performance:

```bash
# Default: 8 parallel workers
zea load "sales/*.parquet"

# Increase workers for large datasets
zea load "sales/*.parquet" --parallel=16

# Reduce workers for limited resources
zea load "sales/*.parquet" --parallel=4
```

**Performance tips:**
- More workers = faster loading (up to a point)
- Diminishing returns beyond CPU core count
- I/O bottleneck may limit gains

### File Limits

Control how many files are loaded:

```bash
# Load maximum 100 files
zea load "sales/*.csv" --max-files=100

# Load maximum 1000 files
zea load "warehouse/**/*.parquet" --max-files=1000

# Unlimited (default)
zea load "sales/*.csv" --max-files=0
```

**Use cases:**
- **Safety**: Prevent accidentally loading too many files
- **Sampling**: Quick analysis of subset
- **Development**: Test on small dataset first

### Format Filtering

Filter files by format before loading:

```bash
# Only CSV files
zea load "sales/**/*" --format=csv

# Only Parquet files
zea load "warehouse/**/*" --format=parquet

# Only JSON files
zea load "data/**/*" --format=json
```

**Supported formats:**
- `csv`
- `tsv`
- `json`
- `jsonl`
- `xml`
- `parquet`

## Schema Evolution

ZeaShell automatically handles files with different schemas:

### Missing Columns

```
# File 1: sales-2026-01.csv
id,amount

# File 2: sales-2026-02.csv
id,amount,region

# Loading both:
zea load "sales-*.csv"
```

**Result:**
- Schema has all columns: `id, amount, region`
- Rows from File 1 have `NULL` for `region`
- Rows from File 2 have all values

### Type Promotion

When column types differ, ZeaShell promotes to most general type:

```
# File 1: amount is int64
id,amount
1,100

# File 2: amount is float64
id,amount
2,99.99

# Result: amount promoted to float64
```

**Type hierarchy:**
1. `int64` → `float64` → `string`
2. Promotions are automatic
3. No data loss

### Schema Preview

Check schema before loading large datasets:

```bash
# Show inferred schema
zea load "sales/" --schema-preview
```

**Output:**
```
Found 156 files

Inferred schema:
  customer: string
  region: string
  amount: int64
  date: string
  product: string
```

**Benefits:**
- No data loading (fast!)
- Verify column names
- Check types before processing
- Confirm file count

## Cloud Storage

ZeaShell works with cloud storage through filesystem mounts (Volumez, FUSE, etc.).

### Volumez Mounts

Mount cloud storage as local filesystem:

```bash
# S3 bucket mounted at /mnt/s3-data
zea load "/mnt/s3-data/sales/date=2026-*/*.parquet"

# Azure Blob mounted at /mnt/azure
zea load "/mnt/azure/warehouse/**/*.csv"

# GCS bucket mounted at /mnt/gcs
zea load "/mnt/gcs/data/*.json"
```

### Multi-Cloud Workflows

```bash
# Load from S3, write to local
zea load "/mnt/s3/sales/*.parquet" \
  | zea filter "amount > 1000" \
  | zea store /mnt/local/high_value.parquet

# Load from Azure, write to S3
zea load "/mnt/azure/raw/*.csv" \
  | zea group region --sum=amount \
  | zea store /mnt/s3/summary.parquet
```

### Benefits

- **Cloud-agnostic**: S3, Azure, GCS, MinIO work identically
- **Standard tooling**: Use `ls`, `find`, `du` on cloud data
- **No SDK dependencies**: ZeaShell doesn't need cloud SDKs
- **Volumez optimizations**: Caching, prefetching, parallel reads

## Real-World Workflows

### Analyze Partitioned Sales Data

```bash
zea load "sales/date=2026-03-*/*.parquet" \
  | zea filter "amount > 1000" \
  | zea group region --sum=amount --count=1 \
  | zea store sales_summary.parquet
```

### Join Partitioned Data with Dimension Tables

```bash
zea load "transactions/date=*/*.parquet" \
  | zea join "dimensions/customers.csv" --on=customer_id \
  | zea filter "tier = 'Gold'" \
  | zea group product --sum=amount
```

### Time-Series Pivot

```bash
zea load "metrics/date=2026-*/*.csv" \
  | zea pivot --index=metric --column=date --values=value \
  | zea store metrics_wide.parquet
```

### Regional Analysis

```bash
# Load West region only
zea load "sales/region=west/*.parquet" \
  | zea group product --sum=amount --count=1 \
  | zea filter "amount_sum > 10000"

# Compare regions
zea load "sales/region=*/*.parquet" \
  | zea group region --sum=amount --avg=amount \
  | zea sort amount_sum:desc
```

### Incremental Loading

```bash
# Load only latest partition
zea load "sales/date=2026-03-17/*.parquet" \
  | zea filter "amount > 0" \
  | zea store daily_summary.csv

# Load last 7 days
zea load "sales/date=2026-03-1[0-7]/*.parquet" \
  | zea group customer --sum=amount
```

### Data Quality Checks

```bash
# Check row counts across partitions
for partition in sales/date=*/; do
  echo "$partition: $(zea load "$partition/*.parquet" | wc -l)"
done

# Verify schema consistency
zea load "sales/" --schema-preview

# Find partitions with anomalies
zea load "sales/date=*/*.parquet" \
  | zea group date --count=1 \
  | zea filter "count < 100"  # Flag low-volume days
```

## Petabyte-Scale Example

```bash
# S3 data lake mounted via Volumez at /mnt/datalake
# 10TB of partitioned Parquet across 1000+ files

zea load "/mnt/datalake/events/year=2026/month=03/**/*.parquet" \
  | zea filter "user_tier = 'premium' AND revenue > 100" \
  | zea group country --sum=revenue --count=1 \
  | zea pivot --index=country --column=product --values=sum_revenue \
  | zea store /mnt/local/premium_revenue_2026_03.parquet
```

**What ZeaShell handles:**
- Parallel loading (1000+ files)
- Schema evolution (files may have different columns)
- Memory-efficient streaming
- Filter pushdown

**What Volumez handles:**
- S3 authentication
- Intelligent caching
- Parallel S3 reads
- Network optimization

## Troubleshooting

### Too Many Files

**Problem**: Loading takes too long or uses too much memory

**Solutions:**
```bash
# Limit file count
zea load "sales/*.csv" --max-files=100

# More specific pattern
zea load "sales/date=2026-03-17/*.csv"  # Instead of sales/*.csv

# Pre-filter partitions
zea load "sales/region=west/date=2026-03-*/*.csv"
```

### Schema Mismatches

**Problem**: Files have incompatible schemas

**Solution:**
```bash
# Preview schema first
zea load "sales/" --schema-preview

# Load compatible subsets
zea load "sales/2026-*.csv"  # Newer files with consistent schema
```

### Performance Issues

**Problem**: Loading is slow

**Solutions:**
```bash
# Increase parallelism
zea load "sales/*.parquet" --parallel=16

# Use Parquet instead of CSV (2-3x faster)
zea load "sales/*.parquet"  # Instead of sales/*.csv

# Load from local cache (via Volumez)
# Volumez caches hot data automatically
```

### Pattern Not Matching

**Problem**: Glob pattern finds no files

**Solutions:**
```bash
# Test with ls first
ls sales/*.csv

# Check quotes (required for glob patterns)
zea load "sales/*.csv"  # ✓ Quoted
zea load sales/*.csv     # ✗ Shell expands before ZeaShell sees it

# Use absolute paths
zea load "/full/path/to/sales/*.csv"
```

## Best Practices

1. **Use schema preview** - Check schema before loading large datasets
2. **Start with limits** - Use `--max-files` when exploring unknown data
3. **Leverage partitions** - Load only needed partitions for faster processing
4. **Tune parallelism** - Match CPU cores for best performance
5. **Prefer Parquet** - 5-10x smaller, 2-3x faster than CSV
6. **Quote patterns** - Always quote glob patterns: `"*.csv"`
7. **Use Volumez mounts** - For cloud storage, mount with Volumez for best performance
8. **Test patterns with ls** - Verify glob patterns with `ls` before using in ZeaShell

## Pattern Cheat Sheet

```bash
# All files in directory
"dir/*"

# All files recursively
"dir/**/*"

# Specific extension
"*.csv"
"dir/**/*.parquet"

# Date partitions
"date=2026-03-*/*.csv"
"date=2026-0[1-3]-*/*.csv"

# Multi-level partitions
"year=2026/month=03/**/*.parquet"

# Multiple patterns
"jan/*.csv,feb/*.csv"

# Complex patterns
"sales/region=*/date=2026-*/*.parquet"
```

## Related Documentation

- [COMMANDS.md](COMMANDS.md) - `zea load` command reference
- [README.md](../README.md) - Quick start and installation
- [examples/glob-demo.sh](../examples/glob-demo.sh) - Live demos of glob patterns
