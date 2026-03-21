# ZeaShell: Local-First Data Engineering from Laptop to Data Lake

**A workflow-first introduction to the terminal analytics engine that bridges exploration and production**

---

## The Problem with Modern Data Tools

Data engineering today often means choosing between two extremes: quick-and-dirty exploration tools (Excel, basic scripts) or full production infrastructure (cloud warehouses, orchestration platforms, authentication layers). The gap between exploring a CSV file on your laptop and running production data pipelines has never been wider.

What if you could learn data engineering patterns with the same instant feedback loop as writing code? What if prototyping a data pipeline didn't require spinning up cloud infrastructure? What if the mental models you learn on your laptop scaled directly to production data lakes?

Enter ZeaShell.

## What is ZeaShell?

ZeaShell is a local-first analytics engine that brings production-grade data processing to your terminal. Think of it as DuckDB meets Unix pipes with an interactive TUI - zero configuration, no cloud setup, your data never leaves your machine.

The core philosophy: **local-first iteration with production-ready patterns**.

You start by exploring a single CSV file interactively. Minutes later, you're running SQL analytics across partitioned Parquet files. An hour later, you're ready to take those exact same patterns to production Spark or Flink pipelines. No context switch. No infrastructure until you need scale.

Let's explore how ZeaShell enables real-world data workflows by composing simple, powerful primitives.

## Basic Data Engineering Workflows: The Building Blocks

Every data pipeline starts with fundamental transforms: filter, select, sort, aggregate, join. ZeaShell implements these as composable commands that work like Unix pipes.

### Filter and Transform

You've just received a sales dataset and need to find high-value transactions in the western region:

```bash
zea load sales.csv | zea filter "amount > 1000 AND region = 'west'"
```

The filter expressions are SQL-like but simplified - no SELECT/FROM boilerplate, just the conditions that matter. [Full expression syntax](https://github.com/open-tempest-labs/zeashell/blob/main/docs/EXPRESSIONS.md).

Need just specific columns? Chain another command:

```bash
zea load sales.csv \
  | zea filter "amount > 1000 AND region = 'west'" \
  | zea select customer,amount,date
```

Each command reads from stdin and writes to stdout. Pure Unix philosophy. This composability is what makes complex workflows manageable.

### Aggregation Pipelines

Sales leadership wants regional summaries. Build an aggregation pipeline:

```bash
zea load sales.csv \
  | zea group region --sum=amount --count=1 --avg=amount \
  | zea sort amount_sum:desc
```

This groups by region, calculates sum/count/average, and sorts by total revenue descending. The output? Clean CSV ready for the next stage.

Now save it:

```bash
zea load sales.csv \
  | zea group region --sum=amount --count=1 --avg=amount \
  | zea sort amount_sum:desc \
  | zea store regional_summary.csv
```

Or save as Parquet for efficient storage:

```bash
zea load sales.csv \
  | zea group region --sum=amount --count=1 --avg=amount \
  | zea store regional_summary.parquet
```

ZeaShell automatically infers the format from the file extension. CSV, JSON, Parquet - all interchangeable. [Format documentation](https://github.com/open-tempest-labs/zeashell/blob/main/docs/COMMANDS.md).

### Joining Datasets

You have sales data and customer data in separate files. Enrich sales with customer information:

```bash
zea load sales.csv \
  | zea join customers.csv --on=customer_id --type=left \
  | zea filter "tier = 'Gold'" \
  | zea group customer_name --sum=amount
```

This performs a left join on customer_id, filters to Gold tier customers, and aggregates their total spend. The join types (inner, left, right, full outer) work exactly like SQL - because they should.

These primitives - filter, select, group, join, sort - compose into sophisticated pipelines. But what about when you don't know your data yet?

## Interactive Data Exploration: See Your Data

Before building pipelines, you need to understand your data. Enter the interactive viewer.

```bash
zea view sales.csv
```

This launches a full-screen terminal UI with your data in a scrollable table. Arrow keys navigate, columns resize automatically, and the status bar shows your position (Row: 247/5,000 | Col: 3/8).

### Exploration Workflows

**See the schema**: Press `d` to inspect column types, null counts, and data quality metrics:

```
┌─ Schema & Metadata ─────────────────────────────┐
│ Schema:                                          │
│   customer_id    int64                           │
│   amount         float64                         │
│   region         string      (12 nulls, 2.4%)   │
│   date           string                          │
│                                                  │
│ Total Rows: 5,000                                │
│ Total Columns: 8                                 │
└──────────────────────────────────────────────────┘
```

Immediately you spot data quality issues - 2.4% of regions are missing.

**Interactive filtering**: Press `f` and enter a filter expression. The view updates instantly showing only matching rows. The status bar shows "Filter: amount > 1000 | Rows: 342/5,000" - you can see the filter's impact in real-time.

**Sort by column**: Navigate to any column and press `s`. The data re-sorts, cycling through ascending, descending, and original order. Perfect for finding outliers.

**View distributions**: Press `g` on a column to see a histogram (for numbers) or bar chart (for categories). You discover that 70% of sales come from just two regions.

**Export filtered data**: After filtering and sorting, press `e` to save the current view. Your exploration becomes a reusable dataset.

The viewer isn't just for looking - it's an interactive query builder. Filter, sort, explore, then press `e` to export. That exported data becomes the input to your next pipeline.

[Full viewer documentation](https://github.com/open-tempest-labs/zeashell/blob/main/docs/VIEWER.md).

## Advanced Analytics with SQL: DuckDB Integration

Sometimes you need SQL's full power. ZeaShell integrates DuckDB for complex analytics while maintaining pipe composability.

### SQL in Pipelines

Your data is already flowing through pipes. Just add SQL:

```bash
zea load sales.csv | zea sql "
  SELECT
    region,
    SUM(amount) as total_revenue,
    COUNT(*) as transaction_count,
    AVG(amount) as avg_transaction
  FROM stdin
  GROUP BY region
  ORDER BY total_revenue DESC
"
```

The input data is available as the `stdin` table. Query it like any SQL table, and the results flow to stdout as CSV.

### Window Functions and Ranking

Find the top 3 transactions per region:

```bash
zea load sales.csv | zea sql "
  WITH ranked AS (
    SELECT
      *,
      ROW_NUMBER() OVER (
        PARTITION BY region
        ORDER BY amount DESC
      ) as rank
    FROM stdin
  )
  SELECT * FROM ranked WHERE rank <= 3
"
```

CTEs, window functions, complex joins - the full DuckDB SQL engine is available.

### Hybrid Workflows

The power comes from mixing paradigms. Use DataFrame operations for simple transforms, SQL for complex analytics:

```bash
# Pre-filter with fast DataFrame ops
zea load sales.csv \
  | zea filter "date >= '2026-01-01'" \
  | zea select customer_id,amount,region,product \
  | zea sql "
    SELECT
      customer_id,
      region,
      SUM(amount) as total_spend,
      COUNT(DISTINCT product) as product_diversity,
      AVG(amount) as avg_transaction,
      STDDEV(amount) as spend_volatility
    FROM stdin
    GROUP BY customer_id, region
    HAVING total_spend > 5000
  " \
  | zea view
```

Filter and select with zea commands (fast, simple), then use SQL for statistical aggregations (powerful, expressive). The result opens in the viewer for interactive exploration.

[SQL architecture and performance](https://github.com/open-tempest-labs/zeashell/blob/main/docs/SQL_ARCHITECTURE.md).

## Production-Scale Patterns: Cloud Storage and Partitioned Data

Local exploration is great, but real data lives in cloud storage and spans thousands of partitioned files. ZeaShell handles this through FUSE-based mounts (like Volumez) and glob patterns.

### Working with Cloud Storage

With Volumez or similar FUSE mounts, cloud storage appears as local directories:

```bash
# S3 bucket mounted at /mnt/s3-data via Volumez
zea load /mnt/s3-data/sales/sales.csv | zea describe
```

To ZeaShell, there's no difference between local and cloud files. The FUSE mount handles the networking. Your data pipelines work identically whether processing local CSV files or petabyte-scale S3 data lakes.

Write results back to cloud storage:

```bash
zea load /mnt/s3-data/raw/sales.csv \
  | zea filter "amount > 1000" \
  | zea group customer_id --sum=amount \
  | zea store /mnt/s3-data/processed/high_value_customers.parquet
```

The exact same commands. The exact same mental model. Local-first development, cloud-scale production.

### Partitioned Data with Glob Patterns

Production data lakes use Hive-style partitioning. ZeaShell handles this natively with glob patterns:

```bash
# Load all March 2026 sales across all partitions
zea load "sales/year=2026/month=03/**/*.parquet" | zea describe
```

The glob pattern matches hundreds or thousands of files. ZeaShell loads them in parallel, merges schemas automatically (handling missing columns gracefully), and presents unified data.

Process data for a specific date range:

```bash
zea load "sales/year=2026/month=*/day=*/sales.parquet" \
  | zea filter "amount > 500" \
  | zea group region,product --sum=amount --count=1 \
  | zea store "analytics/sales_summary.parquet"
```

This reads partitioned Parquet files from a data lake, filters, aggregates, and writes results - all with streaming I/O and parallel processing.

Preview schemas without loading data:

```bash
zea load "sales/**/*.parquet" --schema-preview
```

This scans partitions to show the unified schema without loading actual data. Perfect for understanding unfamiliar data lakes.

[Partitioned data guide](https://github.com/open-tempest-labs/zeashell/blob/main/docs/PARTITIONED_DATA.md).

## Advanced Analytics: Pivot Tables and Complex Transformations

Real analysis often requires reshaping data. ZeaShell provides pivot/unpivot operations that integrate seamlessly with pipelines.

### Pivot: Long to Wide

You have daily sales data and want a matrix with products as rows and dates as columns:

```bash
zea load daily_sales.csv \
  | zea pivot --index=product --column=date --values=amount \
  | zea view
```

This transforms:
```
product,date,amount
Widget,2026-03-01,100
Widget,2026-03-02,150
Gadget,2026-03-01,200
Gadget,2026-03-02,175
```

Into:
```
product,2026-03-01,2026-03-02
Widget,100,150
Gadget,200,175
```

Perfect for spreadsheet-style analysis or time series visualization.

### Unpivot: Wide to Long

The reverse operation - convert wide-format data back to long format for analysis:

```bash
zea load quarterly_sales.csv \
  | zea unpivot --index=region --columns=Q1,Q2,Q3,Q4 \
    --var-name=quarter --value-name=revenue
```

This enables aggregations across previously separate columns.

### Multi-Dimensional Analytics

Combine pivots with SQL for sophisticated analysis:

```bash
# Analyze sales by region and product with time series
zea load "sales/2026/**/*.parquet" \
  | zea sql "
    SELECT
      region,
      product,
      date,
      SUM(amount) as daily_revenue
    FROM stdin
    GROUP BY region, product, date
  " \
  | zea pivot --index=region,product --column=date --values=daily_revenue \
  | zea store regional_product_timeseries.csv
```

SQL for aggregation, pivot for reshaping, store for persistence. Each stage does one thing well.

## Putting It All Together: Real-World Pipeline

Let's build a complete analytics pipeline combining everything we've covered:

**Scenario**: Analyze customer lifetime value from a partitioned data lake, enrich with customer demographics, identify high-value segments, and output pivot tables for leadership.

```bash
# Step 1: Load partitioned sales data, filter recent activity
zea load "/mnt/s3-data/sales/year=2026/**/*.parquet" \
  | zea filter "date >= '2026-01-01'" \
  | zea select customer_id,amount,product,date,region \
  > /tmp/recent_sales.csv

# Step 2: Calculate customer lifetime value with SQL
zea load /tmp/recent_sales.csv \
  | zea sql "
    SELECT
      customer_id,
      region,
      COUNT(*) as transaction_count,
      SUM(amount) as lifetime_value,
      AVG(amount) as avg_transaction,
      COUNT(DISTINCT product) as product_diversity,
      MIN(date) as first_purchase,
      MAX(date) as last_purchase
    FROM stdin
    GROUP BY customer_id, region
  " \
  > /tmp/customer_ltv.csv

# Step 3: Enrich with customer demographics
zea load /tmp/customer_ltv.csv \
  | zea join /mnt/s3-data/customers/demographics.csv \
    --on=customer_id --type=left \
  > /tmp/enriched_customers.csv

# Step 4: Interactive exploration to find segments
zea view /tmp/enriched_customers.csv
# (In viewer: filter by "lifetime_value > 5000", sort by transaction_count,
#  press 'd' to check null rates in demographics, press 'e' to export)

# Step 5: Generate pivot table for regional LTV analysis
zea load /tmp/enriched_customers.csv \
  | zea filter "lifetime_value > 5000" \
  | zea pivot --index=region --column=tier --values=lifetime_value \
  | zea store /mnt/s3-data/analytics/regional_high_value.csv

# Step 6: Advanced segment analysis with SQL
zea load /tmp/enriched_customers.csv \
  | zea sql "
    WITH segments AS (
      SELECT
        *,
        CASE
          WHEN lifetime_value > 10000 THEN 'VIP'
          WHEN lifetime_value > 5000 THEN 'High Value'
          WHEN lifetime_value > 2000 THEN 'Medium Value'
          ELSE 'Standard'
        END as segment
      FROM stdin
    )
    SELECT
      segment,
      region,
      COUNT(*) as customer_count,
      SUM(lifetime_value) as total_revenue,
      AVG(product_diversity) as avg_products
    FROM segments
    GROUP BY segment, region
    ORDER BY total_revenue DESC
  " \
  | zea view
```

This pipeline:
1. Loads partitioned cloud data via FUSE mount
2. Uses DataFrame operations for fast filtering/selection
3. Employs SQL for complex aggregations
4. Enriches with joins from reference data
5. Leverages interactive viewer for exploratory analysis
6. Generates pivot tables for reporting
7. Performs advanced segmentation with window logic

Every step uses standard Unix pipes. Every intermediate result is CSV. Every stage is debuggable, testable, and understandable.

## The Path to Production

Here's where ZeaShell's philosophy pays off: **these patterns scale directly to production**.

The pipeline above translates nearly 1:1 to production systems:

- `zea load` → Spark DataFrame readers or dbt sources
- `zea filter` → WHERE clauses or Spark filters
- `zea group` → GROUP BY or Spark aggregations
- `zea join` → JOIN operations in SQL/Spark/Flink
- `zea sql` → dbt models or Spark SQL
- `zea pivot` → Pandas/Spark pivot operations
- `zea store` → Data warehouse writes or Delta Lake

You learn the mental models locally with instant iteration. When you need scale, you already know the patterns. The main difference? You add orchestration (Airflow, Dagster) and change the execution engine (Spark, Flink, dbt). The data transformations themselves? Identical concepts.

## Why Local-First Matters

ZeaShell's local-first approach isn't just about avoiding cloud costs. It's about the iteration speed that enables learning.

**Instant feedback**: No waiting for cloud instances, no authentication flows, no network latency. Write a pipeline, see results in seconds.

**No infrastructure barrier**: Students, hobbyists, or anyone learning data engineering can start immediately. No credit card, no setup, no YAML hell.

**Privacy by design**: Your data never leaves your machine until you explicitly write it somewhere. Explore sensitive data without cloud security concerns.

**Works offline**: Airplane, coffee shop, anywhere. Your analytics engine travels with you.

**Production patterns**: The workflows you build locally use the same primitives as production data platforms. No toy examples that don't transfer.

This is the onramp data engineering has been missing.

## Getting Started

Install ZeaShell in seconds:

```bash
# Homebrew (macOS/Linux)
brew install open-tempest-labs/zeashell/zeashell

# Go
go install github.com/open-tempest-labs/zeashell/cmd/zea@latest

# From source
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell && go build -o zea ./cmd/zea
```

Try your first pipeline:

```bash
# Get sample data
curl -O https://raw.githubusercontent.com/open-tempest-labs/zeashell/main/examples/sales.csv

# Explore interactively
zea view sales.csv

# Build a pipeline
zea load sales.csv \
  | zea filter "amount > 500" \
  | zea group region --sum=amount --count=1 \
  | zea sort amount_sum:desc
```

## Learn More

- **[GitHub Repository](https://github.com/open-tempest-labs/zeashell)** - Source code and issues
- **[Documentation](https://github.com/open-tempest-labs/zeashell/tree/main/docs)** - Complete guides
  - [Commands Reference](https://github.com/open-tempest-labs/zeashell/blob/main/docs/COMMANDS.md)
  - [Expression Language](https://github.com/open-tempest-labs/zeashell/blob/main/docs/EXPRESSIONS.md)
  - [Interactive Viewer](https://github.com/open-tempest-labs/zeashell/blob/main/docs/VIEWER.md)
  - [Partitioned Data](https://github.com/open-tempest-labs/zeashell/blob/main/docs/PARTITIONED_DATA.md)
  - [SQL Architecture](https://github.com/open-tempest-labs/zeashell/blob/main/docs/SQL_ARCHITECTURE.md)
  - [Plugin System](https://github.com/open-tempest-labs/zeashell/blob/main/docs/PLUGINS.md)

## Conclusion: Local-First, Production-Ready

ZeaShell proves that powerful data engineering doesn't require infrastructure. By bringing production-grade capabilities to your local machine with zero configuration, it creates a new category: the local-first analytics engine.

Start with exploring a CSV file. Graduate to SQL analytics on partitioned data lakes. Take proven patterns to production Spark and dbt pipelines. All with the same mental models, just different scales.

Your data stays local. Your iteration stays instant. Your learning stays focused.

That's the power of local-first data engineering.

---

**Ready to start?** `brew install open-tempest-labs/zeashell/zeashell`

**Questions or feedback?** Open an issue on [GitHub](https://github.com/open-tempest-labs/zeashell/issues)

**Built with:** Go, DuckDB, Apache Arrow, and a love for Unix pipes 🚀
