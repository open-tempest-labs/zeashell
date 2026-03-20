# ZeaShell v0.3.0 - Plugin System & Enhanced Viewer

**DataFrame Shell - CSV to petabytes, one pipe at a time**

This release introduces a powerful plugin system that makes ZeaShell infinitely extensible, plus viewer enhancements for working with truncated cell values.

## 🎉 Major Features

### Plugin System

Extend ZeaShell with custom commands using simple executable scripts. Drop any executable into `~/.zea/plugins/` and it becomes a native `zea run <plugin>` subcommand.

**Key capabilities:**
- 🔌 **Simple installation** - Drop executable scripts into `~/.zea/plugins/`
- 🚀 **Native integration** - Plugins appear as `zea run <name>` commands
- 📝 **Metadata directives** - Use `@name`, `@desc`, `@args` comments for help text
- 🌐 **Language agnostic** - Bash, Python, Ruby, compiled binaries - anything executable works
- 🔄 **Pipeline compatible** - Full stdin/stdout/stderr passthrough
- ✨ **Auto-completion** - Plugins appear in tab completion automatically
- 🐛 **Debug mode** - Use `ZEA_DEBUG=1` to see plugin discovery

**Example plugin:**
```bash
#!/bin/bash
# @desc Process sales CSV files with standard transforms
# @args [file] [min-amount]

zea load "$1" | zea filter "amount > ${2:-1000}" | zea view
```

**Usage:**
```bash
# List available plugins
zea run --help

# Run a plugin
zea run sales-pipeline sales.csv 1000

# Use in pipelines
zea run inventory | zea load --format csv | zea filter "quantity > 100"
```

**Plugin discovery:**
- Default location: `~/.zea/plugins/`
- Custom location: `export ZEA_PLUGINS=/opt/my-plugins`
- Auto-creates directory if missing
- Graceful failure on permission errors

### Enhanced Interactive Viewer

The TUI viewer now allows viewing complete cell values when they're truncated.

**New feature:**
- **Press Enter** on any cell to view its full value in a modal
- Works for long strings, JSON, or any truncated data
- Scrollable modal for very long values
- Press ESC or 'q' to close

**Before:** Long values were truncated with `...`
**Now:** Press Enter to see the complete value in a popup

## 📦 Installation

### Homebrew (macOS/Linux)
```bash
brew tap open-tempest-labs/zeashell
brew install zeashell
# Or upgrade:
brew upgrade zeashell
```

### Using Go Install
```bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@v0.3.0
```

### From Source
```bash
git clone https://github.com/open-tempest-labs/zeashell
cd zeashell
git checkout v0.3.0
go build -o zea ./cmd/zea
```

## 🚀 Quick Start

### Using Plugins

```bash
# Create a plugin
mkdir -p ~/.zea/plugins

cat > ~/.zea/plugins/high-value << 'EOF'
#!/bin/bash
# @desc Find high-value transactions
zea load "$1" | zea filter "amount > 1000" | zea group customer --sum=amount
EOF

chmod +x ~/.zea/plugins/high-value

# Use it
zea run high-value sales.csv
```

### Viewer Enhancement

```bash
# Open a file with long text fields
zea view data.csv

# Navigate to a truncated cell (shows "Some very long te...")
# Press Enter to see the full value in a modal
# Press ESC to close the modal
```

## 📚 Documentation

- **[Plugin System Guide](docs/PLUGINS.md)** - Complete plugin documentation
- **[Plugin Examples](examples/plugins/)** - Ready-to-use example plugins
- **[Commands Reference](docs/COMMANDS.md)** - All ZeaShell commands
- **[Viewer Guide](docs/VIEWER.md)** - Interactive TUI documentation

### New Example Plugins

The release includes 5 ready-to-use example plugins in `examples/plugins/`:

1. **`inventory`** - Generate sample inventory data
2. **`sales-pipeline`** - Process sales with configurable thresholds
3. **`high-value-customers`** - Customer analysis tool
4. **`csv-preview`** - Quick file preview with schema
5. **`json-to-csv`** - Format converter

Copy them to `~/.zea/plugins/` to try them out!

## 🔄 Upgrading

### Homebrew
```bash
brew update
brew upgrade zeashell
```

### Go Install
```bash
go install github.com/open-tempest-labs/zeashell/cmd/zea@v0.3.0
```

## 📝 Complete Feature List

All features from previous releases plus new additions:

**Core Features:**
- 🔄 Pipeable Commands - Full Unix pipe compatibility
- 📊 Interactive TUI - Terminal viewer with sort, filter, graph
- 🚀 Multi-Format - CSV, TSV, JSON, JSONL, XML, Parquet
- 🗂️ Partitioned Data - Glob patterns, parallel multi-file loading
- 🌐 HTTP/HTTPS - Load data directly from URLs
- 🔗 Relational Joins - Inner, left, right, full outer joins
- ↔️ Pivot/Unpivot - Transform between long and wide formats
- 🔌 **NEW: Plugin System** - Extend with custom commands
- ⚡ Fast - Single static binary, columnar storage
- 🎯 Expressive - SQL-like filter expressions
- 🏗️ Production Ready - Type inference, error handling, streaming I/O

**Viewer Enhancements:**
- ✨ **NEW: View full cell values** - Press Enter on truncated cells
- Sort, filter, graph, and export
- Navigation with arrow keys
- Interactive data exploration

## 🐛 Bug Fixes

- Improved error handling in plugin discovery
- Better graceful failure when plugin directory is inaccessible

## 🙏 Contributors

Thanks to all contributors who made this release possible!

## 🔗 Links

- **GitHub**: https://github.com/open-tempest-labs/zeashell
- **Issues**: https://github.com/open-tempest-labs/zeashell/issues
- **Full Changelog**: https://github.com/open-tempest-labs/zeashell/compare/v0.2.0...v0.3.0

---

**Start building ZeaShell plugins today!** 🔌🐚
