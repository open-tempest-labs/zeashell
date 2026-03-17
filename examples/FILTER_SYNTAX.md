# ZeaShell Filter Syntax Guide

## Quick Reference

### For the test data (sales with: customer, region, amount, product)

```bash
# Valid filter expressions:
amount > 1000
amount >= 1500
amount < 2000
region = 'West'
region != 'East'
customer = 'Alice'
product = 'Laptop'

# Combined with AND/OR:
amount > 1000 AND region = 'West'
amount > 2000 OR amount < 800
region = 'West' AND product = 'Server'
```

## Supported Operators

### Comparison Operators
- `=` - Equal to
- `!=` - Not equal to
- `>` - Greater than
- `>=` - Greater than or equal
- `<` - Less than
- `<=` - Less than or equal

### Logical Operators
- `AND` - Both conditions must be true
- `OR` - Either condition must be true

### Special Operators
- `CONTAINS` - For array membership or pattern matching

## Syntax Rules

### 1. String Values - Use Single Quotes
```
✓ region = 'West'
✓ customer = 'Alice'
✗ region = West       (missing quotes)
✗ region = "West"     (use single quotes, not double)
```

### 2. Numeric Values - No Quotes
```
✓ amount > 1000
✓ amount >= 1500
✓ amount = 2200
✗ amount > '1000'     (don't quote numbers)
```

### 3. Column Names - No Quotes
```
✓ amount > 1000
✓ region = 'West'
✗ 'amount' > 1000     (don't quote column names)
```

### 4. Logical Operators - Must Be Uppercase
```
✓ amount > 1000 AND region = 'West'
✗ amount > 1000 and region = 'West'  (lowercase won't work)
✗ amount > 1000 && region = 'West'   (use AND, not &&)
```

## Examples for Test Data

Based on the sales data with columns: `customer, region, amount, product`

### Simple Filters

```bash
# Numeric comparisons
amount > 1000          # Sales over $1000
amount >= 1500         # Sales $1500 or more
amount < 2000          # Sales under $2000
amount = 950           # Exact amount

# String comparisons
region = 'West'        # Only West region
region != 'East'       # All except East
customer = 'Alice'     # Only Alice's sales
product = 'Laptop'     # Only laptop sales
```

### Combined Filters

```bash
# High-value West sales
amount > 1500 AND region = 'West'

# High or low sales (exclude middle)
amount > 3000 OR amount < 800

# Servers in any region except South
product = 'Server' AND region != 'South'

# Multiple conditions
amount > 1000 AND region = 'West' AND product != 'Mouse'
```

### Common Use Cases

```bash
# Top performers
amount > 2000

# Specific regions
region = 'West' OR region = 'East'

# Product focus
product = 'Server' OR product = 'Workstation'

# Value range
amount >= 1000 AND amount <= 3000

# Exclude outliers
amount > 500 AND amount < 4000

# Regional analysis
region = 'West' AND amount > 1000
```

## Troubleshooting

### Filter Not Working?

**Problem:** "No results" or error message

**Common Mistakes:**

1. **Forgot quotes around strings**
   ```
   ✗ region = West
   ✓ region = 'West'
   ```

2. **Wrong case for logical operators**
   ```
   ✗ amount > 1000 and region = 'West'
   ✓ amount > 1000 AND region = 'West'
   ```

3. **Quoted numbers**
   ```
   ✗ amount > '1000'
   ✓ amount > 1000
   ```

4. **Case sensitivity in values**
   ```
   ✗ region = 'west'   (if data has 'West')
   ✓ region = 'West'
   ```

5. **Column name typos**
   ```
   ✗ amounts > 1000    (column is 'amount')
   ✓ amount > 1000
   ```

### Check Your Data First

Before filtering, verify:
- Exact column names (case-sensitive)
- Exact string values (case-sensitive)
- Numeric ranges in your data

In `zea view`:
1. Look at the table header for exact column names
2. Scroll through data to see actual values
3. Use those exact names/values in your filter

## Interactive Testing

### Step-by-Step in zea view

1. **Start viewer:**
   ```bash
   zea view "testdata/sales-partitioned/date=*/*.csv"
   ```

2. **Try simple filters:**
   - Press `f`
   - Type: `amount > 1000`
   - Press Enter
   - Result: Should show 7 rows

3. **Try combined filters:**
   - Press `f`
   - Type: `amount > 1000 AND region = 'West'`
   - Press Enter
   - Result: Should show 4 rows (Frank, Charlie, Ivy, and one more)

4. **Reset and try again:**
   - Press `r` to reset
   - Press `f`
   - Type: `region = 'West' OR region = 'East'`
   - Press Enter

## Advanced Features

### Pattern Matching (CONTAINS)
```bash
# For array columns or pattern matching
tags CONTAINS 'premium'
services CONTAINS 'api.*.prod'
```

### Nested Fields (JSON/XML)
```bash
address.city = 'SF'
address.state = 'CA' AND tags CONTAINS 'premium'
```

### Array Indexing
```bash
orders[0] > 1000
items[1] = 'laptop'
```

## Quick Copy-Paste Examples

For quick testing in `zea view`, just copy these:

```
amount > 1000
amount > 1500 AND region = 'West'
region = 'West'
product = 'Server'
amount < 1000 OR amount > 3000
region != 'South'
```

## Still Not Working?

1. Check the status bar for error messages
2. Press `r` to reset all filters
3. Try the simplest filter first: `amount > 0`
4. If that works, gradually add complexity
5. Make sure column names match exactly (check table header)
