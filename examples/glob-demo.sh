#!/bin/bash
# ZeaShell Glob and Multi-File Loading Demo
# Demonstrates partitioned data loading, glob patterns, and directory operations

set -e

echo "=== ZeaShell Glob Loading Demo ==="
echo ""

# Ensure we're in the right directory
cd "$(dirname "$0")/.."

# Build if needed
if [ ! -f "./zea" ]; then
    echo "Building zea..."
    go build -o zea ./cmd/zea
fi

echo "1. Load entire directory (all files recursively)"
echo "   Command: ./zea load \"testdata/sales-partitioned/\""
echo ""
./zea load "testdata/sales-partitioned/" 2>&1 | head -10
echo "   ... (showing first 10 rows)"
echo ""

echo "2. Schema preview (infer schema without loading)"
echo "   Command: ./zea load \"testdata/sales-partitioned/\" --schema-preview"
echo ""
./zea load "testdata/sales-partitioned/" --schema-preview
echo ""

echo "3. Load with glob pattern (partitioned sales only)"
echo "   Command: ./zea load \"testdata/sales-partitioned/date=*/*.csv\""
echo ""
./zea load "testdata/sales-partitioned/date=*/*.csv" 2>&1 | head -10
echo "   ... (showing first 10 rows)"
echo ""

echo "4. Load specific partitions"
echo "   Command: ./zea load \"testdata/sales-partitioned/date=2026-03-01/*.csv\""
echo ""
./zea load "testdata/sales-partitioned/date=2026-03-01/*.csv" 2>&1
echo ""

echo "5. Full pipeline: Load, Filter, Group, Store"
echo "   Command: ./zea load \"testdata/sales-partitioned/date=*/*.csv\" | ./zea filter \"amount > 1000\" | ./zea group region --sum=amount --count=1"
echo ""
./zea load "testdata/sales-partitioned/date=*/*.csv" 2>/dev/null \
  | ./zea filter "amount > 1000" \
  | ./zea group region --sum=amount --count=1
echo ""

echo "6. Join partitioned data with customers"
echo "   Command: ./zea load \"testdata/sales-partitioned/date=*/*.csv\" | ./zea join testdata/sales-partitioned/customers.csv --on=customer"
echo ""
./zea load "testdata/sales-partitioned/date=*/*.csv" 2>/dev/null \
  | ./zea join testdata/sales-partitioned/customers.csv --on=customer \
  | head -10
echo "   ... (showing first 10 rows)"
echo ""

echo "7. Load and pivot by region"
echo "   Command: ./zea load \"testdata/sales-partitioned/date=*/*.csv\" | ./zea group region,product --sum=amount | ./zea pivot --index=product --column=region --values=sum_amount"
echo ""
./zea load "testdata/sales-partitioned/date=*/*.csv" 2>/dev/null \
  | ./zea group region,product --sum=amount \
  | ./zea pivot --index=product --column=region --values=sum_amount
echo ""

echo "8. Performance test: Load with limited parallelism"
echo "   Command: ./zea load \"testdata/sales-partitioned/\" --parallel=2 --max-files=10"
echo ""
./zea load "testdata/sales-partitioned/" --parallel=2 --max-files=10 2>&1 | head -5
echo "   ... (showing first 5 rows)"
echo ""

echo "=== Demo Complete ==="
echo ""
echo "Key Features Demonstrated:"
echo "  ✓ Directory loading (recursive)"
echo "  ✓ Glob patterns (*.csv, date=*/*.csv)"
echo "  ✓ Schema preview"
echo "  ✓ Parallel multi-file loading"
echo "  ✓ Integration with filter, group, join, pivot"
echo "  ✓ Performance tuning (--parallel, --max-files)"
echo ""
