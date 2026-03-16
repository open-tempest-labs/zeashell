#!/bin/bash

# ZeaShell Parquet Pipeline Demo
# Demonstrates full Parquet read/write support alongside CSV/TSV/JSONL

echo "=== ZeaShell Parquet Pipeline Demo ==="
echo ""

# Example 1: Convert CSV to Parquet
echo "1. Converting CSV to Parquet:"
./zea load examples/sales.csv | ./zea store examples/sales.parquet
echo ""

# Example 2: Load and display Parquet
echo "2. Loading Parquet file:"
./zea load examples/sales.parquet | head -n 6
echo ""

# Example 3: Filter Parquet data
echo "3. Filtering Parquet data (amount > 1000):"
./zea load examples/sales.parquet | ./zea filter "amount > 1000" | head -n 6
echo ""

# Example 4: Parquet to Parquet pipeline
echo "4. Parquet to Parquet transformation:"
./zea load examples/sales.parquet | \
  ./zea filter "region = 'west'" | \
  ./zea store examples/west-sales.parquet
./zea load examples/west-sales.parquet | wc -l
echo ""

# Example 5: Group aggregation on Parquet
echo "5. Grouping Parquet data by region:"
./zea load examples/sales.parquet | \
  ./zea group region --sum=amount --count=1
echo ""

# Example 6: Cross-format pipeline (Parquet → CSV)
echo "6. Parquet to CSV conversion:"
./zea load examples/sales.parquet | \
  ./zea filter "amount > 500" | \
  ./zea select customer,product,amount | \
  ./zea store /tmp/high_value.csv
head -n 6 /tmp/high_value.csv
echo ""

# Example 7: Complex multi-format pipeline
echo "7. Complex pipeline (CSV → filter → Parquet → group → CSV):"
./zea load examples/sales.csv | \
  ./zea filter "region = 'west' OR region = 'east'" | \
  ./zea store /tmp/filtered.parquet

./zea load /tmp/filtered.parquet | \
  ./zea group region,product --sum=amount --count=1 | \
  ./zea filter "amount_sum > 100" | \
  ./zea store /tmp/summary.csv

echo "Results saved to /tmp/summary.csv"
cat /tmp/summary.csv
echo ""

# Example 8: Parquet schema inspection
echo "8. Inspecting Parquet schema:"
./zea load examples/sales.parquet | ./zea describe 2>&1 | head -n 15
echo ""

# Example 9: Performance test - large dataset
echo "9. Performance test - processing Parquet:"
time (./zea load examples/sales.parquet | \
  ./zea filter "amount > 50" | \
  ./zea group customer --sum=amount --count=1 --avg=amount | \
  ./zea store /tmp/customer_stats.parquet) 2>&1
echo ""

# Example 10: Format agnostic pipeline
echo "10. Format-agnostic processing (auto-detection):"
echo "CSV input:"
./zea load examples/sales.csv | ./zea filter "product = 'laptop'" | wc -l
echo "Parquet input (same result):"
./zea load examples/sales.parquet | ./zea filter "product = 'laptop'" | wc -l
echo ""

# Cleanup
echo "Cleaning up temporary files..."
rm -f /tmp/high_value.csv /tmp/filtered.parquet /tmp/summary.csv /tmp/customer_stats.parquet
rm -f examples/west-sales.parquet

echo ""
echo "=== Parquet Demo Complete ==="
echo "ZeaShell now supports:"
echo "  ✓ Parquet read/write"
echo "  ✓ Format auto-detection"
echo "  ✓ Cross-format pipelines"
echo "  ✓ Full Unix pipe compatibility"
echo ""
echo "Try: ./zea load data.parquet | ./zea filter \"expr\" | ./zea store output.parquet"
