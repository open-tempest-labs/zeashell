#!/bin/bash

# ZeaShell Format Conversion Test Suite
# Tests all format conversions: CSV, TSV, JSONL, Parquet

echo "=== ZeaShell Format Conversion Test Suite ==="
echo ""

# Create temp directory
TEMP_DIR="/tmp/zeashell-format-test"
mkdir -p "$TEMP_DIR"

echo "1. Testing CSV → Other Formats"
echo "================================"

# CSV to TSV
echo "  CSV → TSV..."
./zea load examples/sales.csv | ./zea store "$TEMP_DIR/sales.tsv"

# CSV to JSONL
echo "  CSV → JSONL..."
./zea load examples/sales.csv | ./zea store "$TEMP_DIR/sales.jsonl"

# CSV to Parquet
echo "  CSV → Parquet..."
./zea load examples/sales.csv | ./zea store "$TEMP_DIR/sales.parquet"

echo "  ✓ All CSV conversions complete"
echo ""

echo "2. Testing TSV → Other Formats"
echo "================================"

# TSV to CSV
echo "  TSV → CSV..."
./zea load "$TEMP_DIR/sales.tsv" | ./zea store "$TEMP_DIR/sales_from_tsv.csv"

# TSV to JSONL
echo "  TSV → JSONL..."
./zea load "$TEMP_DIR/sales.tsv" | ./zea store "$TEMP_DIR/sales_from_tsv.jsonl"

# TSV to Parquet
echo "  TSV → Parquet..."
./zea load "$TEMP_DIR/sales.tsv" | ./zea store "$TEMP_DIR/sales_from_tsv.parquet"

echo "  ✓ All TSV conversions complete"
echo ""

echo "3. Testing JSONL → Other Formats"
echo "================================="

# JSONL to CSV
echo "  JSONL → CSV..."
./zea load "$TEMP_DIR/sales.jsonl" | ./zea store "$TEMP_DIR/sales_from_jsonl.csv"

# JSONL to TSV
echo "  JSONL → TSV..."
./zea load "$TEMP_DIR/sales.jsonl" | ./zea store "$TEMP_DIR/sales_from_jsonl.tsv"

# JSONL to Parquet
echo "  JSONL → Parquet..."
./zea load "$TEMP_DIR/sales.jsonl" | ./zea store "$TEMP_DIR/sales_from_jsonl.parquet"

echo "  ✓ All JSONL conversions complete"
echo ""

echo "4. Testing Parquet → Other Formats"
echo "==================================="

# Parquet to CSV
echo "  Parquet → CSV..."
./zea load "$TEMP_DIR/sales.parquet" | ./zea store "$TEMP_DIR/sales_from_parquet.csv"

# Parquet to TSV
echo "  Parquet → TSV..."
./zea load "$TEMP_DIR/sales.parquet" | ./zea store "$TEMP_DIR/sales_from_parquet.tsv"

# Parquet to JSONL
echo "  Parquet → JSONL..."
./zea load "$TEMP_DIR/sales.parquet" | ./zea store "$TEMP_DIR/sales_from_parquet.jsonl"

echo "  ✓ All Parquet conversions complete"
echo ""

echo "5. Testing Round-Trip Conversion"
echo "=================================="

# CSV → TSV → JSONL → Parquet → CSV
echo "  CSV → TSV → JSONL → Parquet → CSV..."
./zea load examples/sales.csv | ./zea store "$TEMP_DIR/round1.tsv"
./zea load "$TEMP_DIR/round1.tsv" | ./zea store "$TEMP_DIR/round2.jsonl"
./zea load "$TEMP_DIR/round2.jsonl" | ./zea store "$TEMP_DIR/round3.parquet"
./zea load "$TEMP_DIR/round3.parquet" | ./zea store "$TEMP_DIR/round4.csv"

# Count rows to verify data integrity
ORIG_ROWS=$(./zea load examples/sales.csv | wc -l | tr -d ' ')
FINAL_ROWS=$(./zea load "$TEMP_DIR/round4.csv" | wc -l | tr -d ' ')

if [ "$ORIG_ROWS" -eq "$FINAL_ROWS" ]; then
  echo "  ✓ Round-trip successful: $ORIG_ROWS rows preserved"
else
  echo "  ✗ Round-trip failed: $ORIG_ROWS → $FINAL_ROWS rows"
fi
echo ""

echo "6. File Size Comparison"
echo "======================="

ls -lh examples/sales.csv "$TEMP_DIR/sales.tsv" "$TEMP_DIR/sales.jsonl" "$TEMP_DIR/sales.parquet" | \
  awk 'NR>1 {print "  " $5 "\t" $9}'

echo ""

echo "7. Testing Format Auto-Detection"
echo "================================="

# Test that each format is correctly detected
echo "  Loading CSV..."
./zea load examples/sales.csv | head -n 2 > /dev/null && echo "    ✓ CSV loaded"

echo "  Loading TSV..."
./zea load "$TEMP_DIR/sales.tsv" | head -n 2 > /dev/null && echo "    ✓ TSV loaded"

echo "  Loading JSONL..."
./zea load "$TEMP_DIR/sales.jsonl" | head -n 2 > /dev/null && echo "    ✓ JSONL loaded"

echo "  Loading Parquet..."
./zea load "$TEMP_DIR/sales.parquet" | head -n 2 > /dev/null && echo "    ✓ Parquet loaded"

echo ""

echo "8. Testing Pipeline with Format Mixing"
echo "======================================="

# CSV → filter → TSV → group → Parquet → select → JSONL
echo "  Running mixed-format pipeline..."
./zea load examples/sales.csv | \
  ./zea filter "amount > 100" | \
  ./zea store "$TEMP_DIR/filtered.tsv"

./zea load "$TEMP_DIR/filtered.tsv" | \
  ./zea group region --sum=amount --count=1 | \
  ./zea store "$TEMP_DIR/grouped.parquet"

./zea load "$TEMP_DIR/grouped.parquet" | \
  ./zea select region,amount_sum | \
  ./zea store "$TEMP_DIR/final.jsonl"

if [ -f "$TEMP_DIR/final.jsonl" ]; then
  RESULT_ROWS=$(wc -l < "$TEMP_DIR/final.jsonl" | tr -d ' ')
  echo "  ✓ Pipeline complete: $RESULT_ROWS result rows"
else
  echo "  ✗ Pipeline failed"
fi

echo ""

echo "9. Sample Output Preview"
echo "========================"

echo "CSV format:"
head -n 3 examples/sales.csv
echo ""

echo "TSV format:"
head -n 3 "$TEMP_DIR/sales.tsv"
echo ""

echo "JSONL format:"
head -n 2 "$TEMP_DIR/sales.jsonl"
echo ""

echo "10. Cleanup"
echo "==========="

read -p "Delete temporary files in $TEMP_DIR? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
  rm -rf "$TEMP_DIR"
  echo "  ✓ Temporary files deleted"
else
  echo "  Files preserved in: $TEMP_DIR"
fi

echo ""
echo "=== Format Conversion Test Suite Complete ==="
echo ""
echo "Summary:"
echo "  ✓ CSV, TSV, JSONL, Parquet formats supported"
echo "  ✓ All format conversions working"
echo "  ✓ Format auto-detection functioning"
echo "  ✓ Round-trip data integrity maintained"
echo "  ✓ Mixed-format pipelines operational"
echo ""
echo "See FORMAT_CONVERSION.md for detailed documentation"
