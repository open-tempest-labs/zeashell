# ZeaShell Expression Language

Complete reference for filter expressions used in `zea filter` and `zea view`.

## Overview

ZeaShell supports a powerful expression language for filtering data with **path-based column names** from flattened JSON and XML structures.

## Operators

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Equal to | `amount = 100` |
| `!=` | Not equal to | `region != 'South'` |
| `>` | Greater than | `amount > 1000` |
| `>=` | Greater than or equal | `amount >= 500` |
| `<` | Less than | `amount < 2000` |
| `<=` | Less than or equal | `amount <= 1500` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `AND` | Logical AND (both conditions must be true) | `amount > 100 AND region = 'West'` |
| `OR` | Logical OR (either condition must be true) | `amount < 500 OR amount > 3000` |

**Important**: Logical operators MUST be uppercase (`AND`, `OR`). Lowercase (`and`, `or`) will not work.

### Special Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `CONTAINS` | Array membership or wildcard pattern matching | `tags CONTAINS 'premium'` |
|  |  | `name CONTAINS '*.webshell.*'` |

## Value Types

### Numbers

Numbers are written without quotes:
```
amount > 1000
price = 99.99
count >= 5
```

Supported:
- Integers: `100`, `-50`, `0`
- Decimals: `99.99`, `3.14`, `-2.5`

### Strings

Strings are written with single quotes:
```
region = 'West'
customer = 'Alice'
name != ''
```

**Important**: Use single quotes (`'`), not double quotes (`"`).

### Booleans

Boolean literals (for boolean columns):
```
active = true
deleted = false
```

## Path-Based Column Names

ZeaShell automatically flattens nested JSON and XML structures into dotted column names:

### JSON Flattening

```json
{
  "customer": "Alice",
  "address": {
    "city": "SF",
    "state": "CA"
  }
}
```

Becomes columns:
- `customer`
- `address.city`
- `address.state`

**Filter expressions:**
```bash
address.city = 'SF'
address.state = 'CA' AND customer = 'Alice'
```

### XML Flattening

```xml
<topology>
  <gateway>
    <provider>
      <role>authentication</role>
    </provider>
  </gateway>
  <service>
    <role>WEBHDFS</role>
  </service>
</topology>
```

Becomes columns:
- `gateway.provider.role`
- `service.role`

**Filter expressions:**
```bash
gateway.provider.role = 'authentication'
service.role = 'WEBHDFS'
```

### Benefits

- **No metadata pollution**: No format-specific columns like `_element`
- **Natural filtering**: Use dotted paths directly
- **Unified model**: Same semantics across JSON and XML
- **Path semantics**: Column names preserve hierarchical structure

## Pattern Matching with CONTAINS

The `CONTAINS` operator supports:
1. **Array membership** - Check if array contains a value
2. **Wildcard patterns** - Match strings with `*` and `?`

### Array Membership

For array-valued columns:
```bash
# Check if array contains value
tags CONTAINS 'premium'
orders CONTAINS 1005
```

### Wildcard Patterns

Wildcard pattern matching supports:
- `*` - Match any number of characters (including zero)
- `?` - Match exactly one character

Examples:
```bash
# Match any name with 'webshell' in the middle
name CONTAINS '*.webshell.*'

# Match service roles starting with 'WEB'
service.role CONTAINS 'WEB*'

# Match server names like server-01, server-02
host CONTAINS 'server-??'

# Match production APIs
service CONTAINS '*.prod.*'
```

## Expression Syntax

### Simple Comparisons

```bash
# Numeric comparisons
amount > 100
amount >= 1000
count < 50
price = 99.99

# String comparisons
region = 'West'
customer != ''
status = 'active'

# Path-based columns
address.city = 'SF'
service.role = 'WEBHDFS'
gateway.provider.name = 'ShiroProvider'
```

### Logical Combinations

```bash
# AND - both conditions must be true
amount > 100 AND region = 'West'
price >= 10 AND price <= 100
customer != '' AND status = 'active'

# OR - either condition must be true
region = 'West' OR region = 'East'
amount < 50 OR amount > 1000
status = 'pending' OR status = 'review'

# Complex combinations (precedence: AND before OR)
amount > 100 AND region = 'West' OR amount > 5000
# Equivalent to: (amount > 100 AND region = 'West') OR (amount > 5000)
```

### Pattern Matching

```bash
# Exact match in arrays
tags CONTAINS 'premium'
categories CONTAINS 'electronics'

# Wildcard patterns
name CONTAINS '*.webshell.*'
host CONTAINS 'server-*'
service.role CONTAINS 'WEB*'
path CONTAINS '*.prod.?????'
```

## Examples by Use Case

### Numeric Filtering

```bash
# Range queries
amount > 1000
amount >= 500 AND amount <= 2000

# Excluding values
amount != 0
price > 0

# Multiple thresholds
amount < 100 OR amount > 5000
```

### String Filtering

```bash
# Exact match
region = 'West'
status = 'active'

# Not equal
customer != ''
region != 'South'

# Multiple values
region = 'West' OR region = 'East'
status = 'pending' OR status = 'approved'
```

### Path-Based Filtering (JSON/XML)

```bash
# JSON paths
address.city = 'SF'
address.state = 'CA' AND address.zip = '94105'
user.profile.tier = 'Gold'

# XML paths
gateway.provider.role = 'authentication'
service.role = 'WEBHDFS'
topology.gateway.name = 'knox'

# Nested structures
config.database.host = 'localhost'
config.database.port > 5432
```

### Pattern Matching

```bash
# Service discovery
service.name CONTAINS '*.prod.*'
service.role CONTAINS 'WEB*'

# Log analysis
message CONTAINS '*.error.*'
path CONTAINS '*.webshell.*'

# Host filtering
hostname CONTAINS 'server-??'
fqdn CONTAINS '*.example.com'
```

### Complex Queries

```bash
# High-value West region sales
amount > 1000 AND region = 'West'

# Exclude outliers
amount > 100 AND amount < 10000

# Multi-region analysis
(region = 'West' OR region = 'East') AND amount > 500

# Path-based with pattern
address.city = 'SF' AND tags CONTAINS 'premium'

# Service filtering
service.role CONTAINS 'WEB*' AND service.url CONTAINS 'localhost'
```

## Syntax Rules

### Column Names

- **NO quotes**: `amount > 100` ✓
- Use dotted paths directly: `address.city = 'SF'` ✓
- Case-sensitive: `Region` ≠ `region`

### String Values

- **Single quotes required**: `region = 'West'` ✓
- Not double quotes: `region = "West"` ✗
- Empty strings: `customer != ''` ✓

### Numbers

- **NO quotes**: `amount > 1000` ✓
- Not quoted: `amount > '1000'` ✗
- Decimals allowed: `price = 99.99` ✓

### Logical Operators

- **UPPERCASE required**: `AND`, `OR` ✓
- Not lowercase: `and`, `or` ✗
- Not symbols: `&&`, `||` ✗

## Operator Precedence

Operators are evaluated in this order:
1. Comparison operators (`=`, `!=`, `>`, `<`, etc.)
2. `AND`
3. `OR`

Examples:
```bash
# This expression:
amount > 100 AND region = 'West' OR amount > 5000

# Is equivalent to:
(amount > 100 AND region = 'West') OR (amount > 5000)

# To force different grouping, use separate filters:
zea filter "amount > 100 AND region = 'West'" | zea filter "amount > 5000"
```

## Type Coercion

ZeaShell automatically handles type conversions:

### Numeric Comparisons

```bash
# String to number
"100" > 50        # Compares as numbers: true

# Int to float
10 > 9.5          # Compares as floats: true
```

### String Comparisons

```bash
# Number to string (fallback)
amount = '100'    # Tries numeric first, then string
```

## NULL Handling

NULL values in data:
- Comparisons with NULL return false
- `!= NULL` pattern doesn't work as expected

To handle missing data, filter before NULL-containing operations.

## Error Messages

Common error messages and solutions:

### "invalid expression"

**Cause**: Parser couldn't understand the expression

**Solutions**:
- Check operator spelling: `AND` not `and`
- Check quotes: `'West'` not `West`
- Check column name exists

### "column not found"

**Cause**: Referenced column doesn't exist

**Solutions**:
- Check exact column name (case-sensitive)
- Use `zea describe` to see available columns
- For path-based columns, check exact path

### "cannot parse value"

**Cause**: Value format is invalid

**Solutions**:
- Strings need quotes: `'value'`
- Numbers don't need quotes: `1000`
- Check for typos

## Best Practices

1. **Start simple** - Test single conditions before combining
2. **Check column names** - Use `zea describe` or `zea view` to see exact names
3. **Use path-based columns** - Leverage dotted notation for JSON/XML
4. **Uppercase logical operators** - Always use `AND`, `OR`
5. **Quote strings, not numbers** - `amount > 100 AND region = 'West'`
6. **Test interactively** - Use `zea view` with `f` key to test expressions
7. **Build incrementally** - Add conditions one at a time

## Related Documentation

- [COMMANDS.md](COMMANDS.md) - `zea filter` and `zea view` commands
- [VIEWER.md](VIEWER.md) - Interactive filtering in the viewer
- [examples/FILTER_SYNTAX.md](../examples/FILTER_SYNTAX.md) - Syntax examples and troubleshooting
