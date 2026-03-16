# ZeaShell Quick Start Guide

Welcome to ZeaShell! This guide will get you up and running in minutes.

## Build

```bash
go build -o zea ./cmd/zea
```

## Basic Usage

### 1. Load Data

```bash
./zea load examples/sales.csv
```

### 2. Filter Rows

```bash
./zea load examples/sales.csv | ./zea filter "amount > 100"
```

### 3. Select Columns

```bash
./zea load examples/sales.csv | ./zea select customer,amount
```

### 4. Group and Aggregate

```bash
# Sum by region
./zea load examples/sales.csv | ./zea group region --sum=amount

# Count by product
./zea load examples/sales.csv | ./zea group product --count=1

# Multiple aggregations
./zea load examples/sales.csv | ./zea group product --sum=amount --count=1 --avg=amount
```

### 5. Save Results

```bash
./zea load examples/sales.csv | ./zea filter "amount > 1000" | ./zea store output.csv
```

### 6. Describe Data

```bash
./zea load examples/sales.csv | ./zea describe
```

## Complete Pipeline Example

Find top customers by total sales:

```bash
./zea load examples/sales.csv | \
  ./zea filter "amount > 50" | \
  ./zea select customer,amount | \
  ./zea group customer --sum=amount --count=1 | \
  ./zea filter "amount_sum > 1000"
```

## Running the Demo

```bash
./examples/sales-pipeline.sh
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Explore more examples in the `examples/` directory
- Check the expression language syntax for advanced filtering
- Try different aggregation functions (sum, avg, min, max, count)

Happy data processing! 🌊
