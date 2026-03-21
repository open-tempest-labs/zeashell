#!/bin/bash
# ZeaShell + DuckDB SQL Integration Demo
# Demonstrates hybrid pipelines mixing DataFrame operations and SQL

set -e

echo "=== ZeaShell + DuckDB SQL Integration Demo ==="
echo ""

# Test 1: Simple aggregation
echo "1. Simple SQL aggregation:"
echo "   Command: zea load sales.csv | zea sql \"SELECT region, SUM(amount) as total FROM stdin GROUP BY region\""
echo ""
zea load sales.csv | zea sql "SELECT region, SUM(amount) as total FROM stdin GROUP BY region"
echo ""

# Test 2: Window functions
echo "2. Window functions for ranking:"
echo "   Command: zea load sales.csv | zea sql \"SELECT customer, amount, ROW_NUMBER() OVER (ORDER BY amount DESC) as rank FROM stdin\""
echo ""
zea load sales.csv | zea sql "SELECT customer, amount, ROW_NUMBER() OVER (ORDER BY amount DESC) as rank FROM stdin" | head -8
echo "   ... (truncated)"
echo ""

# Test 3: Hybrid pipeline (filter + SQL)
echo "3. Hybrid pipeline (DataFrame filter + SQL aggregation):"
echo "   Command: zea load sales.csv | zea filter \"amount > 500\" | zea sql \"SELECT customer, COUNT(*) as count, SUM(amount) as total FROM stdin GROUP BY customer\""
echo ""
zea load sales.csv | zea filter "amount > 500" | zea sql "SELECT customer, COUNT(*) as count, SUM(amount) as total FROM stdin GROUP BY customer ORDER BY total DESC"
echo ""

# Test 4: Subquery
echo "4. Subquery (amounts above average):"
echo "   Command: zea load sales.csv | zea sql \"SELECT customer, region, amount FROM stdin WHERE amount > (SELECT AVG(amount) FROM stdin)\""
echo ""
zea load sales.csv | zea sql "SELECT customer, region, amount FROM stdin WHERE amount > (SELECT AVG(amount) FROM stdin)"
echo ""

# Test 5: CTE (Common Table Expression)
echo "5. CTE for complex analytics:"
echo "   Command: zea load sales.csv | zea sql \"WITH regional_stats AS (...)\""
echo ""
zea load sales.csv | zea sql "WITH regional_stats AS (SELECT region, SUM(amount) as total FROM stdin GROUP BY region) SELECT region, total, ROUND(total * 100.0 / SUM(total) OVER (), 2) as pct FROM regional_stats"
echo ""

# Test 6: String operations
echo "6. String operations and formatting:"
echo "   Command: zea load sales.csv | zea sql \"SELECT UPPER(customer) as customer, CONCAT('$', ROUND(amount, 2)) as formatted_amount FROM stdin\""
echo ""
zea load sales.csv | zea sql "SELECT UPPER(customer) as customer, CONCAT('$', ROUND(amount, 2)) as formatted_amount FROM stdin" | head -8
echo "   ... (truncated)"
echo ""

# Test 7: Query without stdin
echo "7. SQL query without input data:"
echo "   Command: zea sql \"SELECT 1 as x, 2 as y, 'test' as z\""
echo ""
zea sql "SELECT 1 as x, 2 as y, 'test' as z"
echo ""

# Test 8: SQL to view
echo "8. SQL results in interactive viewer:"
echo "   Command: zea load sales.csv | zea sql \"SELECT region, SUM(amount) as total FROM stdin GROUP BY region\" | zea describe"
echo ""
zea load sales.csv | zea sql "SELECT region, SUM(amount) as total FROM stdin GROUP BY region" | zea describe
echo ""

# Test 9: Multiple aggregations
echo "9. Multiple aggregations in one query:"
echo "   Command: zea load sales.csv | zea sql \"SELECT region, COUNT(*) as transactions, SUM(amount) as total, AVG(amount) as avg, MAX(amount) as max FROM stdin GROUP BY region\""
echo ""
zea load sales.csv | zea sql "SELECT region, COUNT(*) as transactions, SUM(amount) as total, ROUND(AVG(amount), 2) as avg, MAX(amount) as max FROM stdin GROUP BY region"
echo ""

# Test 10: Pipeline composition
echo "10. Full pipeline: Load → Filter → SQL → Select → Store:"
echo "    Command: zea load sales.csv | zea filter \"region = 'west'\" | zea sql \"SELECT customer, SUM(amount) as total FROM stdin GROUP BY customer\" | zea select customer,total | zea store /tmp/west_summary.csv"
echo ""
zea load sales.csv | zea filter "region = 'west'" | zea sql "SELECT customer, SUM(amount) as total FROM stdin GROUP BY customer" | zea select customer,total | zea store /tmp/west_summary.csv
echo "Saved to: /tmp/west_summary.csv"
cat /tmp/west_summary.csv
rm /tmp/west_summary.csv
echo ""

echo "=== Demo Complete ==="
echo ""
echo "Key Takeaways:"
echo "  • DuckDB SQL integrates seamlessly with ZeaShell pipelines"
echo "  • Mix DataFrame operations (filter, select) with SQL for flexibility"
echo "  • Full SQL support: aggregations, CTEs, window functions, subqueries"
echo "  • stdin table automatically available for SQL queries"
echo "  • Results are CSV-formatted for continued pipeline processing"
