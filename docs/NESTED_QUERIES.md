# Nested Field Queries: True PICK-Style Multi-Valued Support

ZeaShell provides **true PICK OS-inspired multi-valued field queries** through nested JSON field operations. This allows you to query directly into arrays and nested objects without flattening your data.

## Quick Reference

| Operation | Syntax | Example |
|-----------|--------|---------|
| **Implicit ANY** | `field op value` (on arrays) | `orders > 1000` |
| **Array Contains** | `field CONTAINS value` | `orders CONTAINS 1005` |
| **Array Indexing** | `field[index] op value` | `orders[0] > 1000` |
| **Nested Path** | `field.path op value` | `address.city = 'SF'` |

### Implicit ANY Semantics

When you use comparison operators (`>`, `>=`, `<`, `<=`) on array fields **without indexing**, ZeaShell automatically checks if **ANY element** in the array satisfies the condition. This provides intuitive PICK-like behavior:

```bash
# These are equivalent:
zea load data.json | zea filter "orders > 1000"      # ANY element > 1000
zea load data.json | zea filter "orders CONTAINS >1000"  # (conceptually)

# If you need specific position, use indexing:
zea load data.json | zea filter "orders[0] > 1000"  # First element > 1000
```

## Implicit ANY: Natural Array Comparisons

When comparing arrays with scalars using `>`, `>=`, `<`, `<=`, ZeaShell automatically checks if **ANY element** satisfies the condition. This provides the most intuitive behavior for multi-valued fields.

### Syntax
```
field operator value    # on array field, checks ANY element
```

### Examples

```bash
# Find customers with ANY order > 1000
zea load data.json | zea filter "orders > 1000"

# Find customers with ANY order < 1002
zea load data.json | zea filter "orders < 1002"

# Find products with ANY price >= 50
zea load products.json | zea filter "prices >= 50"
```

### PICK OS Equivalence

```pick
* PICK OS (implied ANY semantics)
SELECT CUSTOMERS WITH ORDERS > 1000

# ZeaShell equivalent - same natural syntax!
zea load customers.json | zea filter "orders > 1000"
```

### How It Works

```bash
# Data: {"customer": "Alice", "orders": [1001, 1002, 1003]}
# Query: orders > 1002
# Result: MATCH (because 1003 > 1002)

# Data: {"customer": "Bob", "orders": [1005, 1006]}
# Query: orders > 1002
# Result: MATCH (because both 1005 and 1006 > 1002)

# Data: {"customer": "Charlie", "orders": [900, 950]}
# Query: orders > 1002
# Result: NO MATCH (no element > 1002)
```

## Array Contains: PICK Multi-Values

The `CONTAINS` operator provides **exact value matching** within JSON arrays. Use this when you need to check for a specific value, not a range.

### Syntax
```
field CONTAINS value
```

### Examples

```bash
# Find customers with order 1005 exactly
zea load data.json | zea filter "orders CONTAINS 1005"

# Find premium customers (exact tag match)
zea load data.json | zea filter "tags CONTAINS 'premium'"

# Combined with other conditions
zea load data.json | zea filter "region = 'west' AND tags CONTAINS 'vip'"
```

### Implicit ANY vs CONTAINS

```bash
# Use implicit ANY for ranges:
zea load data.json | zea filter "orders > 1000"       # ANY order > 1000

# Use CONTAINS for exact matches:
zea load data.json | zea filter "orders CONTAINS 1005"  # Has order 1005 exactly
```

### PICK OS Equivalence

```pick
* PICK OS
SELECT CUSTOMERS WITH TAGS CONTAINING "PREMIUM"

# ZeaShell equivalent
zea load customers.json | zea filter "tags CONTAINS 'premium'"
```

## Array Indexing: Access Specific Elements

Extract and compare specific elements from JSON arrays by index (0-based).

### Syntax
```
field[index] operator value
```

### Examples

```bash
# First order greater than 1000
zea load data.json | zea filter "orders[0] > 1000"

# Second order equals 1005
zea load data.json | zea filter "orders[1] = 1005"

# Third tag is not 'urgent'
zea load data.json | zea filter "tags[2] != 'urgent'"
```

### PICK OS Equivalence

```pick
* PICK OS (accessing multi-valued field position)
SELECT CUSTOMERS WITH ORDERS<1> > 1000

# ZeaShell equivalent
zea load customers.json | zea filter "orders[0] > 1000"
```

## Nested Path Navigation: Sub-Values

Navigate into nested JSON objects using dot notation, similar to PICK's sub-valued fields.

### Syntax
```
field.subfield.property operator value
```

### Examples

```bash
# Filter by city
zea load data.json | zea filter "address.city = 'SF'"

# Filter by state
zea load data.json | zea filter "address.state = 'CA'"

# Deeply nested
zea load data.json | zea filter "metadata.shipping.carrier = 'UPS'"
```

### PICK OS Equivalence

```pick
* PICK OS (sub-valued fields)
SELECT CUSTOMERS WITH ADDRESS.CITY = "SF"

# ZeaShell equivalent
zea load customers.json | zea filter "address.city = 'SF'"
```

## Combined Queries

Combine multiple nested operations with `AND`/`OR` for complex PICK-style queries.

### Examples

```bash
# California customers with specific order
zea load data.json | zea filter "address.state = 'CA' AND orders CONTAINS 1005"

# Multiple conditions
zea load data.json | zea filter "tags CONTAINS 'vip' AND orders[0] > 1000"

# Complex logic
zea load data.json | zea filter "address.city = 'SF' OR (address.state = 'CA' AND tags CONTAINS 'premium')"
```

## Complete Example: E-Commerce Pipeline

```bash
# Create sample data
cat > customers.json <<'EOF'
[
  {
    "customer_id": 1001,
    "name": "Alice",
    "address": {
      "city": "San Francisco",
      "state": "CA",
      "zip": "94102"
    },
    "orders": [5001, 5002, 5003],
    "tags": ["vip", "premium", "wholesale"],
    "metadata": {
      "discount": 0.15,
      "shipping": "express"
    }
  },
  {
    "customer_id": 1002,
    "name": "Bob",
    "address": {
      "city": "Oakland",
      "state": "CA",
      "zip": "94601"
    },
    "orders": [5010, 5011],
    "tags": ["standard"],
    "metadata": {
      "discount": 0.0,
      "shipping": "standard"
    }
  },
  {
    "customer_id": 1003,
    "name": "Charlie",
    "address": {
      "city": "Portland",
      "state": "OR",
      "zip": "97201"
    },
    "orders": [6001, 6002, 6003, 6004],
    "tags": ["vip", "bulk"],
    "metadata": {
      "discount": 0.20,
      "shipping": "freight"
    }
  }
]
EOF

# Example queries
echo "=== VIP customers in California ==="
zea load customers.json | \
  zea filter "address.state = 'CA' AND tags CONTAINS 'vip'" | \
  zea select name,address,tags

echo "=== Customers with more than 3 orders ==="
zea load customers.json | \
  zea filter "orders[3] != ''" | \
  zea select name,orders

echo "=== Premium customers with express shipping ==="
zea load customers.json | \
  zea filter "tags CONTAINS 'premium' AND metadata.shipping = 'express'" | \
  zea select name,metadata

echo "=== West coast VIP customers ==="
zea load customers.json | \
  zea filter "(address.state = 'CA' OR address.state = 'OR') AND tags CONTAINS 'vip'" | \
  zea group address.state --count=1
```

## Operator Support Matrix

All standard comparison operators work with nested field access:

| Operator | Array Index | Nested Path | Array Contains |
|----------|-------------|-------------|----------------|
| `=` | ✅ | ✅ | N/A |
| `!=` | ✅ | ✅ | N/A |
| `>` | ✅ | ✅ | N/A |
| `>=` | ✅ | ✅ | N/A |
| `<` | ✅ | ✅ | N/A |
| `<=` | ✅ | ✅ | N/A |
| `CONTAINS` | N/A | N/A | ✅ |

## Advanced Use Cases

### 1. Customer Segmentation
```bash
# Identify high-value customers in specific regions
zea load customers.json | \
  zea filter "tags CONTAINS 'vip' AND address.state = 'CA'" | \
  zea group address.city --count=1
```

### 2. Order Analysis
```bash
# Find customers whose first order exceeds threshold
zea load customers.json | \
  zea filter "orders[0] > 5000" | \
  zea select name,orders
```

### 3. Tag-Based Filtering
```bash
# Multiple tag requirements
zea load products.json | \
  zea filter "tags CONTAINS 'electronics' AND tags CONTAINS 'premium'" | \
  zea store premium_electronics.json
```

### 4. Nested Metadata Queries
```bash
# Complex metadata filtering
zea load customers.json | \
  zea filter "metadata.discount > 0.10 AND metadata.shipping = 'express'" | \
  zea select name,metadata
```

## Error Handling

### Missing Paths
If a nested path doesn't exist, the row won't match (treated as false):
```bash
# Rows without 'address.city' won't match
zea load data.json | zea filter "address.city = 'SF'"
```

### Array Out of Bounds
Array indices that don't exist won't match (no error thrown):
```bash
# Rows with fewer than 4 orders won't match
zea load data.json | zea filter "orders[3] > 1000"
```

### Non-Array CONTAINS
If a field isn't an array, CONTAINS does direct comparison:
```bash
# Falls back to equality check if not an array
zea load data.json | zea filter "status CONTAINS 'active'"
```

## Best Practices

### ✅ DO: Leverage PICK-Style Multi-Values
```bash
# Store related data together
zea load orders.json | \
  zea filter "items CONTAINS 'laptop' AND tags CONTAINS 'express'" | \
  zea store urgent_laptop_orders.json
```

### ✅ DO: Use Nested Paths for Structured Data
```bash
# Query into complex objects
zea load api_response.json | \
  zea filter "user.profile.verified = true" | \
  zea select user.name,user.email
```

### ✅ DO: Combine Operations
```bash
# Mix different nested operations
zea load data.json | \
  zea filter "orders[0] > 1000 AND address.state = 'CA' AND tags CONTAINS 'premium'"
```

### ⚠️ DON'T: Assume All Rows Have Fields
```bash
# Use != to exclude missing fields
zea load data.json | zea filter "address.city != ''"
```

## Performance Considerations

- **JSON Parsing**: Nested queries parse JSON at query time (slight overhead)
- **Array Scans**: CONTAINS operator scans entire array (O(n) per array)
- **Path Navigation**: Dot notation navigates object tree (fast)
- **Recommendation**: For extremely large datasets, consider Parquet for flat data

## Migration from Workarounds

### Before (Limitations)
```bash
# Had to use exact string match
zea load data.json | zea filter "orders = '[1005,1006]'"

# Or pre-process with jq
jq '.[] | select(.orders[] == 1005)' data.json | zea load
```

### After (Native Support)
```bash
# Direct nested queries!
zea load data.json | zea filter "orders CONTAINS 1005"
zea load data.json | zea filter "orders[0] > 1004"
```

## See Also

- [JSON_SUPPORT.md](JSON_SUPPORT.md) - Complete JSON format guide
- [JSON_COMPLETE.md](JSON_COMPLETE.md) - Implementation summary
- [README.md](README.md) - Main documentation

---

**ZeaShell: True PICK OS multi-valued data processing for the modern age** 🗄️✨
