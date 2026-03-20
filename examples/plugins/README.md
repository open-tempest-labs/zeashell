# ZeaShell Plugin Examples

This directory contains example plugins demonstrating various use cases for extending ZeaShell.

## Installation

Copy any example plugins to your `~/.zea/plugins/` directory:

```bash
# Copy all examples
cp examples/plugins/* ~/.zea/plugins/
chmod +x ~/.zea/plugins/*

# Or copy individual plugins
cp examples/plugins/inventory ~/.zea/plugins/
chmod +x ~/.zea/plugins/inventory
```

## Available Examples

### Data Generators

#### `inventory`
Generate sample inventory data for testing.

```bash
zea run inventory | zea load --format csv | zea view
```

### Pipeline Wrappers

#### `sales-pipeline`
Process sales CSV files with filtering and aggregation.

```bash
zea run sales-pipeline sales.csv 1000
```

Features:
- Configurable minimum amount threshold
- Automatic aggregation by customer
- Sorted by total amount

#### `high-value-customers`
Find customers with total purchases over a threshold.

```bash
zea run high-value-customers sales.csv 5000
```

Output includes:
- Customer name
- Total purchase amount
- Transaction count

### Utilities

#### `csv-preview`
Quick preview of CSV files with schema and sample data.

```bash
zea run csv-preview data.csv
```

Shows:
- File size
- Schema with data types
- Sample rows

#### `json-to-csv`
Convert JSON or JSONL files to CSV format.

```bash
# Convert to file
zea run json-to-csv data.json output.csv

# Output to stdout
zea run json-to-csv data.jsonl | head -10
```

## Creating Your Own Plugins

Use these examples as templates for your own plugins. Key patterns:

### 1. Help Text

Always provide usage information:

```bash
if [ -z "$1" ]; then
    echo "Usage: zea run myplugin <arguments>"
    echo "Description of what the plugin does"
    exit 1
fi
```

### 2. Metadata Directives

Add metadata at the top of your script:

```bash
#!/bin/bash
# @name my-plugin
# @desc Short description
# @args [required] [optional]
```

### 3. Error Handling

Check for required files and dependencies:

```bash
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: File not found: $INPUT_FILE" >&2
    exit 1
fi
```

### 4. Pipeline Integration

Output CSV for easy integration with ZeaShell:

```bash
zea load "$INPUT" | zea filter "amount > 100"
```

## Testing Plugins

Test plugins directly before installing:

```bash
# Test the script directly
./examples/plugins/inventory

# Test with ZeaShell pipeline
./examples/plugins/inventory | zea load --format csv | zea describe
```

## More Information

- [Plugin System Documentation](../../docs/PLUGINS.md)
- [Commands Reference](../../docs/COMMANDS.md)
- [Expression Language](../../docs/EXPRESSIONS.md)

## Contributing

Have a useful plugin? Share it with the community:

1. Add it to `examples/plugins/` with documentation
2. Include usage examples
3. Submit a PR to the repository

---

**Happy plugin building!** 🔌
