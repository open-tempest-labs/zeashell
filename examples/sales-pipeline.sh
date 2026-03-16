#!/bin/bash

# ZeaShell Sales Pipeline Demo
# This script demonstrates the power of ZeaShell for data processing

echo "=== ZeaShell Sales Pipeline Demo ==="
echo ""

# Example 1: Load and describe data
echo "1. Loading sales data and showing schema:"
./zea load examples/sales.csv | ./zea describe
echo ""

# Example 2: Filter high-value sales
echo "2. Filter sales over $100:"
./zea load examples/sales.csv | ./zea filter "amount > 100" | head -n 6
echo ""

# Example 3: Select specific columns
echo "3. Select customer and amount columns:"
./zea load examples/sales.csv | ./zea select customer,amount | head -n 6
echo ""

# Example 4: Group by region and sum amounts
echo "4. Total sales by region:"
./zea load examples/sales.csv | ./zea group region --sum=amount
echo ""

# Example 5: Group by product and calculate statistics
echo "5. Product statistics (sum and count):"
./zea load examples/sales.csv | ./zea group product --sum=amount --count=1
echo ""

# Example 6: Complex pipeline - filter, select, and group
echo "6. West region laptop sales summary:"
./zea load examples/sales.csv | \
  ./zea filter "region = 'west' AND product = 'laptop'" | \
  ./zea select customer,amount | \
  ./zea group customer --sum=amount --count=1
echo ""

# Example 7: Filter and store results
echo "7. Storing high-value sales to file:"
./zea load examples/sales.csv | \
  ./zea filter "amount > 100" | \
  ./zea store examples/high_value_sales.csv
echo ""

# Example 8: Multiple filters with AND
echo "8. West region sales over $1000:"
./zea load examples/sales.csv | \
  ./zea filter "region = 'west' AND amount > 1000"
echo ""

# Example 9: Group by multiple columns
echo "9. Sales by region and product:"
./zea load examples/sales.csv | \
  ./zea group region,product --sum=amount --count=1 | head -n 11
echo ""

# Example 10: Full pipeline - load, filter, select, group, store
echo "10. Complete analysis pipeline (saved to examples/analysis.csv):"
./zea load examples/sales.csv | \
  ./zea filter "amount > 50" | \
  ./zea select customer,region,amount | \
  ./zea group customer,region --sum=amount --count=1 | \
  ./zea store examples/analysis.csv

echo ""
echo "=== Demo Complete ==="
echo "Try your own pipelines with: ./zea load examples/sales.csv | ./zea filter \"your_expression\""
