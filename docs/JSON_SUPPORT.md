# JSON Support: PICK-Style Multi-Valued Data

ZeaShell includes full JSON support, inspired by PICK OS's multi-valued database philosophy. Unlike JSONL (one object per line), full JSON allows for rich nested structures that mirror PICK's hierarchical data model.

## Why JSON (Not Just JSONL)?

**PICK OS Heritage**: PICK databases supported multi-valued fields natively - fields could contain lists of values, and those lists could themselves be multi-valued. JSON's hierarchical structure is the modern equivalent.

**JSON in ZeaShell**:
- Preserves nested objects and arrays
- Multi-valued fields stored as JSON strings
- Round-trip conversion maintains structure
- Perfect for complex, hierarchical data

## Supported JSON Formats

### 1. Array of Objects (Most Common)
```json
[
  {"name": "Alice", "age": 30, "skills": ["Go", "Python"]},
  {"name": "Bob", "age": 25, "skills": ["JavaScript"]}
]
```

### 2. Single Object (Converts to Single-Row Frame)
```json
{
  "name": "Alice",
  "age": 30,
  "skills": ["Go", "Python"],
  "address": {
    "city": "SF",
    "state": "CA"
  }
}
```

### 3. Columnar JSON (Object with Array Values)
```json
{
  "names": ["Alice", "Bob"],
  "ages": [30, 25],
  "skills": [["Go", "Python"], ["JavaScript"]]
}
```

## Multi-Valued Fields: PICK-Style

### Example: Nested Structures

```json
[
  {
    "customer": "Alice",
    "orders": [1001, 1002, 1003],
    "addresses": [
      {"type": "home", "city": "SF"},
      {"type": "work", "city": "Oakland"}
    ],
    "tags": ["premium", "urgent"]
  }
]
```

**Loading this file:**
```bash
zea load complex_data.json
```

**Output (nested structures preserved as JSON strings):**
```
customer,orders,addresses,tags
Alice,"[1001,1002,1003]","[{""type"":""home"",""city"":""SF""},{""type"":""work"",""city"":""Oakland""}]","[""premium"",""urgent""]"
```

The nested structures are preserved! You can process them through pipelines and they remain intact.

## JSON vs JSONL

| Feature | JSON (.json) | JSONL (.jsonl) |
|---------|--------------|----------------|
| Format | Array of objects `[{...},{...}]` | One object per line `{...}\n{...}` |
| Readability | Pretty-printed, indented | Compact, one line each |
| Streaming | Must load entire file | Can stream line-by-line |
| Nested Data | ✅ Fully preserved | ✅ Fully preserved |
| File Size | Larger (formatting) | Smaller (no formatting) |
| Use Case | APIs, config files, rich data | Logs, streaming, large datasets |
| PICK-like | ✅ Yes - hierarchical | Partial |

## Usage Examples

### Example 1: Load JSON with Nested Data

```bash
# Create a JSON file with nested structures
cat > products.json <<'EOF'
[
  {
    "id": 1,
    "name": "Laptop",
    "specs": {
      "cpu": "Intel i7",
      "ram": "16GB",
      "storage": "512GB SSD"
    },
    "tags": ["electronics", "computers", "premium"],
    "prices": {
      "retail": 1200,
      "wholesale": 900
    }
  },
  {
    "id": 2,
    "name": "Mouse",
    "specs": {
      "type": "wireless",
      "dpi": 1600
    },
    "tags": ["electronics", "accessories"],
    "prices": {
      "retail": 25,
      "wholesale": 15
    }
  }
]
EOF

# Load and view
zea load products.json
```

Output:
```
id,name,specs,tags,prices
1,Laptop,"{""cpu"":""Intel i7"",""ram"":""16GB"",""storage"":""512GB SSD""}","[""electronics"",""computers"",""premium""]","{""retail"":1200,""wholesale"":900}"
2,Mouse,"{""type"":""wireless"",""dpi"":1600}","[""electronics"",""accessories""]","{""retail"":25,""wholesale"":15}"
```

### Example 2: Process and Export as JSON

```bash
# Load CSV, process, export as pretty JSON
zea load sales.csv | \
  zea filter "amount > 1000" | \
  zea select customer,product,amount | \
  zea store high_value.json

# Result is pretty-printed JSON
cat high_value.json
```

```json
[
  {
    "amount": 1200.5,
    "customer": "Alice",
    "product": "laptop"
  },
  {
    "amount": 1150,
    "customer": "David",
    "product": "laptop"
  }
]
```

### Example 3: JSON → CSV → JSON Round-Trip

```bash
# Start with JSON
zea load data.json | \
  # Process as CSV in pipeline
  zea filter "price > 100" | \
  zea group category --sum=price | \
  # Export back to JSON
  zea store summary.json
```

Nested structures remain intact through the entire pipeline!

### Example 4: API Response Processing

```bash
# Fetch API response (JSON array)
curl https://api.example.com/users > users.json

# Process with ZeaShell
zea load users.json | \
  zea filter "active = true" | \
  zea select id,name,email | \
  zea store active_users.csv
```

## Multi-Valued Field Operations

### Preserving Complex Structures

```bash
# Load JSON with nested data
zea load complex.json | \
  # Filter still works on top-level fields
  zea filter "customer = 'Alice'" | \
  # Nested structures preserved
  zea store filtered.json
```

### Converting to Flat Structure

```bash
# If you need to flatten, convert to CSV
zea load nested.json | \
  zea store flat.csv

# Nested fields become JSON strings in CSV
```

### Round-Trip with Nesting

```bash
# JSON → Process → JSON (nesting maintained)
zea load api_response.json | \
  zea filter "status = 'active'" | \
  zea group category --count=1 | \
  zea store summary.json
```

## Format Conversions with JSON

### JSON ↔ CSV
```bash
# JSON to CSV (nested becomes strings)
zea load data.json | zea store data.csv

# CSV to JSON (pretty-printed)
zea load data.csv | zea store data.json
```

### JSON ↔ JSONL
```bash
# JSON to JSONL (array to lines)
zea load data.json | zea store data.jsonl

# JSONL to JSON (lines to array)
zea load data.jsonl | zea store data.json
```

### JSON ↔ Parquet
```bash
# JSON to Parquet (efficient storage)
zea load large_data.json | zea store compressed.parquet

# Parquet to JSON (for inspection)
zea load data.parquet | zea store readable.json
```

## PICK OS Comparison

| PICK Concept | ZeaShell JSON Equivalent |
|--------------|--------------------------|
| Multi-valued field | JSON array: `"tags": ["a", "b"]` |
| Sub-valued field | Nested array: `"data": [[1,2],[3,4]]` |
| Dictionary | JSON object: `{"key": "value"}` |
| Association | Nested object: `{"user": {"name": "Alice"}}` |
| Item | JSON object in array |
| File | JSON array of objects |

## Best Practices

### 1. Use JSON for Rich Data
- API responses
- Configuration files
- Data with nested structures
- Human-readable exports

### 2. Use JSONL for Streaming
- Large datasets
- Log files
- Event streams
- Line-by-line processing

### 3. Preserve Structure
```bash
# Good: JSON → process → JSON
zea load api.json | zea filter "..." | zea store result.json

# Structures preserved through pipeline
```

### 4. Flatten When Needed
```bash
# Convert to CSV for flat analysis
zea load nested.json | zea store flat.csv

# Nested fields become quoted JSON strings
```

## Performance Considerations

### JSON File Sizes
- **JSON**: Larger (pretty-printing, indentation)
- **JSONL**: Smaller (no formatting)
- **Parquet**: Smallest (compression)

Example (same data):
```
data.json    →  2.5 MB  (pretty-printed)
data.jsonl   →  1.8 MB  (compact)
data.parquet →  0.3 MB  (compressed)
```

### Processing Speed
- **JSON**: Slower (parse entire file)
- **JSONL**: Faster (line-by-line)
- **Parquet**: Fastest (columnar, compressed)

### Recommendations
1. **Small datasets (<10MB)**: Use JSON for readability
2. **Large datasets (>100MB)**: Convert to Parquet
3. **Streaming/logs**: Use JSONL
4. **APIs/config**: Use JSON

## Advanced Examples

### Multi-Valued PICK-Style Data

```json
[
  {
    "customer_id": "C001",
    "name": "Alice",
    "orders": [
      {
        "order_id": 1001,
        "items": [
          {"sku": "LAP01", "qty": 1, "price": 1200},
          {"sku": "MOU01", "qty": 2, "price": 25}
        ],
        "total": 1250
      },
      {
        "order_id": 1002,
        "items": [
          {"sku": "KEY01", "qty": 1, "price": 89}
        ],
        "total": 89
      }
    ],
    "tags": ["vip", "wholesale", "net30"]
  }
]
```

**Load and process:**
```bash
zea load customers_orders.json | \
  zea select customer_id,name,tags | \
  zea filter "customer_id = 'C001'" | \
  zea store vip_customer.json
```

All nested structures (`orders`, `items`, `tags`) are preserved!

## Nested Field Query Operations

### Array Operations with CONTAINS

**Query into arrays** to find specific values:

```bash
# Find rows where orders array contains 1005
zea load data.json | zea filter "orders CONTAINS 1005"

# Find customers with premium tag
zea load data.json | zea filter "tags CONTAINS 'premium'"
```

**How it works**: The CONTAINS operator parses JSON arrays and checks if the target value exists within them.

### Array Indexing

**Access specific array elements** by index:

```bash
# Find rows where first order is greater than 1004
zea load data.json | zea filter "orders[0] > 1004"

# Find rows where second order equals 1002
zea load data.json | zea filter "orders[1] = 1002"

# Check if third item in list exists and matches
zea load data.json | zea filter "items[2] = 'laptop'"
```

**How it works**: Array indexing extracts the element at the specified position (0-based) and performs the comparison.

### Nested Path Navigation

**Navigate into nested objects** using dot notation:

```bash
# Find customers in San Francisco
zea load data.json | zea filter "address.city = 'SF'"

# Find all California customers
zea load data.json | zea filter "address.state = 'CA'"

# Check nested metadata
zea load data.json | zea filter "metadata.priority = 'high'"
```

**How it works**: Dot notation navigates through nested JSON objects to extract the target field value.

### Combined Nested Queries

**Combine multiple nested operations** with AND/OR:

```bash
# California customers with order 1005
zea load data.json | zea filter "address.state = 'CA' AND orders CONTAINS 1005"

# SF or Portland customers
zea load data.json | zea filter "address.city = 'SF' OR address.city = 'Portland'"

# Complex combination
zea load data.json | zea filter "address.state = 'CA' AND orders[0] > 1000 AND tags CONTAINS 'premium'"
```

### Comprehensive Examples

#### Example 1: E-commerce Orders
```bash
# Find VIP customers in California with high-value orders
cat > orders.json <<'EOF'
[
  {
    "customer": "Alice",
    "address": {"city": "SF", "state": "CA"},
    "orders": [1001, 1002, 1003],
    "tags": ["vip", "premium"]
  },
  {
    "customer": "Bob",
    "address": {"city": "Oakland", "state": "CA"},
    "orders": [1005, 1006],
    "tags": ["standard"]
  },
  {
    "customer": "Charlie",
    "address": {"city": "Portland", "state": "OR"},
    "orders": [2001, 2002],
    "tags": ["vip"]
  }
]
EOF

# Find VIP customers in CA
zea load orders.json | \
  zea filter "address.state = 'CA' AND tags CONTAINS 'vip'"
```

#### Example 2: Product Catalog
```bash
# Query nested product specifications
zea load products.json | \
  zea filter "specs.cpu = 'Intel i7' AND tags CONTAINS 'premium'" | \
  zea select name,specs

# Find products with specific array values
zea load products.json | \
  zea filter "categories[0] = 'electronics'" | \
  zea group category --count=1
```

### Supported Operators for Nested Fields

All standard comparison operators work with nested fields:

| Operator | Nested Path | Array Index | Array Contains |
|----------|-------------|-------------|----------------|
| `=` | ✅ | ✅ | N/A |
| `!=` | ✅ | ✅ | N/A |
| `>` | ✅ | ✅ | N/A |
| `>=` | ✅ | ✅ | N/A |
| `<` | ✅ | ✅ | N/A |
| `<=` | ✅ | ✅ | N/A |
| `CONTAINS` | N/A | N/A | ✅ |

### Best Practices

1. **✅ DO**: Use CONTAINS for array membership
   ```bash
   zea load data.json | zea filter "tags CONTAINS 'urgent'"
   ```

2. **✅ DO**: Use array indexing for specific positions
   ```bash
   zea load data.json | zea filter "orders[0] > 1000"
   ```

3. **✅ DO**: Use dot notation for nested objects
   ```bash
   zea load data.json | zea filter "address.city = 'SF'"
   ```

4. **✅ DO**: Combine with AND/OR for complex queries
   ```bash
   zea load data.json | zea filter "address.state = 'CA' AND orders CONTAINS 1005"
   ```

5. **✅ DO**: Use JSON for PICK-style multi-valued data processing
   ```bash
   # True PICK-style: query into multi-valued fields
   zea load api.json | zea filter "tags CONTAINS 'premium'" | zea store vip.json
   ```

### Error Handling

- **Missing paths**: If a nested path doesn't exist in a row, that row won't match
- **Out of bounds**: Array indices that don't exist won't match (no error)
- **Type mismatches**: Non-array fields with CONTAINS won't match
- **Invalid JSON**: If a field isn't valid JSON, it's treated as a regular string

### When to Use Each Format

- **JSON with nested queries**: PICK-style multi-valued data processing
- **Flat CSV/TSV**: Simple tabular data without nesting
- **JSONL**: Streaming large JSON datasets with nested structures
- **Parquet**: Best query performance on flat data

## See Also

- [FORMAT_CONVERSION.md](FORMAT_CONVERSION.md) - All format conversions
- [PARQUET.md](PARQUET.md) - Parquet documentation
- [README.md](README.md) - Main documentation

---

**ZeaShell JSON: Modern multi-valued data, inspired by PICK OS** 🗄️
