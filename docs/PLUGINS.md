# ZeaShell Plugin System

**Extend ZeaShell with custom commands using simple executable scripts**

The plugin system allows you to create custom `zea run <plugin>` subcommands by dropping executable scripts into `~/.zea/plugins/` (or `$ZEA_PLUGINS`). Plugins integrate seamlessly with ZeaShell's pipeline architecture and appear as native commands.

## Quick Start

### 1. Create a plugin directory

```bash
mkdir -p ~/.zea/plugins
```

### 2. Create an executable script

```bash
cat > ~/.zea/plugins/sales << 'EOF'
#!/bin/bash
# @desc Process sales CSV files with standard transforms
# @args [file] [options]

if [ -z "$1" ]; then
    echo "Usage: zea run sales <file.csv>"
    exit 1
fi

zea load "$1" | zea filter "amount > 1000" | zea view
EOF

chmod +x ~/.zea/plugins/sales
```

### 3. Run your plugin

```bash
zea run sales data.csv
```

That's it! Your plugin now works like any native `zea` command.

## Plugin Metadata

Add special comment directives at the top of your script to customize how it appears in ZeaShell:

```bash
#!/bin/bash
# @name my-plugin        # Override command name (defaults to filename)
# @desc Short description of what this plugin does
# @args [required] [optional]  # Usage hint for arguments
```

### Metadata Directives

| Directive | Description | Example |
|-----------|-------------|---------|
| `@name` | Command name (defaults to filename) | `# @name sales-pipeline` |
| `@desc` | Short description shown in help | `# @desc Process sales data` |
| `@args` | Arguments hint for help text | `# @args [file] [options]` |

**Note:** Only the first 20 lines of the script are scanned for metadata.

## Example Plugins

### Data Generator

Generate sample data that can be piped into ZeaShell:

```bash
#!/bin/bash
# @name inventory
# @desc Generate sample inventory data

cat << DATA
item,quantity,price
Widget A,150,29.99
Widget B,200,39.99
Widget C,75,49.99
DATA
```

**Usage:**
```bash
zea run inventory | zea load --format csv | zea view
```

### Pipeline Wrapper

Create reusable pipeline templates:

```bash
#!/bin/bash
# @name high-value-sales
# @desc Find high-value sales transactions
# @args [input-file]

if [ -z "$1" ]; then
    echo "Usage: zea run high-value-sales <sales.csv>"
    exit 1
fi

zea load "$1" \
  | zea filter "amount > 1000 AND status = 'completed'" \
  | zea group customer --sum=amount --count=1 \
  | zea sort amount_sum:desc
```

**Usage:**
```bash
zea run high-value-sales sales.csv | zea store report.csv
```

### External Tool Integration

Integrate external tools with ZeaShell pipelines:

```bash
#!/bin/bash
# @name k8s-pods
# @desc List Kubernetes pods as CSV
# @args [namespace]

NAMESPACE=${1:-default}

kubectl get pods -n "$NAMESPACE" -o json 2>/dev/null \
  | jq -r '["NAME","STATUS","RESTARTS","AGE"], (.items[] | [.metadata.name, .status.phase, (.status.containerStatuses[0].restartCount // 0), .metadata.creationTimestamp]) | @csv' \
  || echo "kubectl not available"
```

**Usage:**
```bash
zea run k8s-pods production | zea load --format csv | zea filter "RESTARTS > 0"
```

### Multi-Step Processing

Combine multiple operations:

```bash
#!/bin/bash
# @name analyze-logs
# @desc Process and analyze log files
# @args [log-directory]

LOG_DIR=${1:-.}

# Find all log files, extract errors, count by severity
find "$LOG_DIR" -name "*.log" -type f | while read logfile; do
    grep -E "ERROR|WARN|INFO" "$logfile" | \
    awk '{print $1","$2","$3}'
done | zea load --format csv | zea group severity --count=1 | zea view
```

## Plugin Discovery

### Directory Precedence

Plugins are loaded from the first available location:

1. **`$ZEA_PLUGINS`** environment variable (if set)
2. **`~/.zea/plugins/`** (default location)

```bash
# Use custom directory
export ZEA_PLUGINS=/opt/my-zea-plugins
zea run my-plugin

# Use default location
zea run my-plugin  # Loads from ~/.zea/plugins/
```

### Debug Mode

Enable debug output to see plugin discovery details:

```bash
ZEA_DEBUG=1 zea run --help
```

Output:
```
Debug: discovered 4 plugin(s) in /Users/you/.zea/plugins
Debug: registered plugin 'inventory'
Debug: registered plugin 'k8s-pods'
Debug: registered plugin 'sales'
Debug: registered plugin 'test'
```

## Plugin Requirements

### Executable Permissions

Scripts must be executable. ZeaShell skips non-executable files:

```bash
chmod +x ~/.zea/plugins/my-plugin
```

### Shebang Line

Start scripts with appropriate shebang:

```bash
#!/bin/bash          # Bash script
#!/usr/bin/env python3   # Python script
#!/usr/bin/env ruby      # Ruby script
```

### Script Languages

Any language that produces an executable file works:

- **Bash/Shell** - Most common
- **Python** - `#!/usr/bin/env python3`
- **Ruby** - `#!/usr/bin/env ruby`
- **Perl** - `#!/usr/bin/env perl`
- **Compiled binaries** - Go, Rust, C, etc.

## Command Integration

### Listing Plugins

```bash
# Show all available plugins
zea run --help

# Show specific plugin help
zea run sales --help
```

### Auto-completion

Plugins automatically support shell auto-completion:

```bash
zea run <TAB>
# Shows: inventory  k8s-pods  sales  test
```

### Pipeline Compatibility

Plugins work seamlessly in ZeaShell pipelines:

```bash
# Plugin as data source
zea run inventory | zea load --format csv | zea filter "quantity > 100"

# Plugin as processor
zea load data.csv | zea run custom-transform | zea store output.csv

# Plugin in the middle
zea load sales.csv | zea run enrich-data | zea group customer --sum=amount
```

## Best Practices

### 1. Provide Help Text

Always include usage information when called without arguments:

```bash
#!/bin/bash
# @desc Process customer data

if [ -z "$1" ]; then
    echo "Usage: zea run customers <input.csv>"
    echo "Process customer data with standard transforms"
    exit 1
fi

# ... rest of script
```

### 2. Use Meaningful Names

Choose descriptive command names:

```bash
# Good
sales-pipeline
customer-analysis
k8s-health-check

# Avoid
script1
temp
test
```

### 3. Output CSV for Pipelines

When generating data, output CSV for easy integration:

```bash
#!/bin/bash
# @desc Generate test data

echo "id,name,value"
echo "1,Alice,100"
echo "2,Bob,200"
```

### 4. Handle Stdin/Stdout

Respect Unix pipe conventions:

```bash
#!/bin/bash
# @desc Filter and transform data

# Read from stdin if no file specified
INPUT=${1:-/dev/stdin}

zea load "$INPUT" | zea filter "amount > 100" | zea select customer,amount
```

### 5. Error Handling

Provide clear error messages:

```bash
#!/bin/bash
# @desc Load external API data

if ! command -v curl &> /dev/null; then
    echo "Error: curl is required but not installed" >&2
    exit 1
fi

if [ -z "$API_KEY" ]; then
    echo "Error: API_KEY environment variable not set" >&2
    exit 1
fi

# ... rest of script
```

## Troubleshooting

### Plugin Not Showing Up

1. **Check permissions:**
   ```bash
   ls -la ~/.zea/plugins/
   # Should show -rwxr-xr-x (executable)
   ```

2. **Check directory:**
   ```bash
   echo $ZEA_PLUGINS  # If set, plugins must be here
   ls ~/.zea/plugins/  # Default location
   ```

3. **Enable debug mode:**
   ```bash
   ZEA_DEBUG=1 zea run --help
   ```

### Plugin Execution Fails

1. **Test directly:**
   ```bash
   ~/.zea/plugins/my-plugin arg1 arg2
   ```

2. **Check shebang:**
   ```bash
   head -1 ~/.zea/plugins/my-plugin
   # Should be #!/bin/bash or similar
   ```

3. **Verify dependencies:**
   ```bash
   # If script uses jq, kubectl, etc.
   which jq kubectl
   ```

### Name Collision

If a plugin name conflicts with an existing file:

- **Filename takes precedence** over `@name` directive
- Plugin commands are under `zea run <name>`, not `zea <name>`
- Rename the file if needed

## Advanced Usage

### Multi-File Plugins

Create a plugin that wraps multiple scripts:

```bash
#!/bin/bash
# @name data-pipeline
# @desc Complete data processing pipeline

PLUGIN_DIR="$HOME/.zea/plugins/pipeline-scripts"

# Step 1: Extract
"$PLUGIN_DIR/extract.sh" "$1"

# Step 2: Transform
"$PLUGIN_DIR/transform.sh"

# Step 3: Load
"$PLUGIN_DIR/load.sh" "$2"
```

### Configuration Files

Load plugin configuration:

```bash
#!/bin/bash
# @desc Load data from configured source

CONFIG="$HOME/.zea/plugins/config.env"
[ -f "$CONFIG" ] && source "$CONFIG"

DATABASE=${DATABASE:-localhost}
TABLE=${TABLE:-sales}

# Connect and extract data
mysql -h "$DATABASE" -e "SELECT * FROM $TABLE" | zea load --format csv
```

### Sharing Plugins

Create a shared plugin repository for your team:

```bash
# Team shared plugins
export ZEA_PLUGINS=/opt/company/zea-plugins

# Per-user overrides
ln -s /opt/company/zea-plugins/common-* ~/.zea/plugins/
```

## Examples Repository

Complete plugin examples are available in the ZeaShell repository:

```bash
# Clone the repository
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell/examples/plugins/

# Copy example plugins
cp *.sh ~/.zea/plugins/
chmod +x ~/.zea/plugins/*.sh

# Try them out
zea run --help
```

## Plugin Ideas

- **Data generators** - Create test datasets
- **API integrations** - Fetch data from REST APIs
- **Database exports** - Extract from MySQL, PostgreSQL, etc.
- **Cloud resource queries** - AWS, GCP, Azure resource listings
- **Log processors** - Parse and transform log formats
- **Report templates** - Standard analysis pipelines
- **Data validators** - Check data quality and schema
- **Format converters** - Custom format transformations

## Contributing Plugins

Share useful plugins with the community:

1. Create your plugin following best practices
2. Add documentation and examples
3. Submit a PR to the [ZeaShell repository](https://github.com/open-tempest-labs/zeashell)

## See Also

- [Commands Reference](COMMANDS.md) - Built-in ZeaShell commands
- [Expression Language](EXPRESSIONS.md) - Filter expressions for use in plugins
- [Partitioned Data](PARTITIONED_DATA.md) - Working with multi-file datasets
- [README](../README.md) - Main documentation

---

**Happy plugin building!** 🔌
