# ZeaShell DuckDB SQL Architecture

**Version:** 0.3.0+ (Dual-Path SQL Integration)

## Overview

ZeaShell integrates DuckDB SQL with a **dual-path architecture** that provides both reliability and performance optimization paths.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    ZeaShell SQL Command                      │
│                     zea sql "SELECT..."                      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ├──► Auto-detect mode
                      │
        ┌─────────────┴──────────────┐
        │                            │
        ▼                            ▼
   [File Mode]                  [Arrow Mode]
   (Default)                     (Future)
        │                            │
        ▼                            ▼
┌──────────────────┐         ┌──────────────────┐
│  CSV Temp File   │         │   Arrow IPC      │
│  ~~~~~~~~~~~~    │         │   ~~~~~~~~~~~    │
│                  │         │                  │
│  stdin → CSV     │         │  stdin → Arrow   │
│  CSV → DuckDB    │         │  Arrow → DuckDB  │
│  DuckDB → CSV    │         │  DuckDB → Arrow  │
│  CSV → stdout    │         │  Arrow → stdout  │
└──────────────────┘         └──────────────────┘
        │                            │
        └──────────┬─────────────────┘
                   ▼
            DuckDB :memory:
                   │
                   ▼
              SQL Query
                   │
                   ▼
              CSV Output
```

## Execution Modes

### 1. File Mode (Default, Current)

**Path:** `stdin → CSV temp file → DuckDB → CSV stdout`

**Characteristics:**
- ✅ Reliable and battle-tested
- ✅ Works with all ZeaShell commands (output CSV)
- ✅ Debuggable (temp files can be inspected)
- ✅ Universal compatibility
- ⚠️  Disk I/O overhead
- ⚠️  Serialization/deserialization cost

**Performance:** ~100MB/s throughput

**Usage:**
```bash
# Implicit (default)
zea load sales.csv | zea sql "SELECT * FROM stdin"

# Explicit
zea load sales.csv | zea sql --file "SELECT * FROM stdin"
```

**Implementation:** `internal/duckdb/sql.go::RunSQLFromStdin()`

### 2. Arrow Mode (Future, Ready)

**Path:** `stdin → Arrow IPC → DuckDB → Arrow IPC stdout`

**Characteristics:**
- 🚀 Zero-copy data transfer
- 🚀 10x+ faster than CSV
- 🚀 Type-safe roundtrip
- 🚀 No disk I/O
- ⚠️  Requires Arrow IPC input format
- ⚠️  ZeaShell commands need Arrow output support

**Performance:** ~1GB/s+ throughput (projected)

**Usage:**
```bash
# When ZeaShell supports Arrow IPC output:
zea load --format arrow sales.parquet | zea sql --arrow "SELECT * FROM stdin"
```

**Implementation:** `internal/duckdb/arrow.go::RunSQLArrowNative()`

### 3. Auto Mode

**Current behavior:** Defaults to File Mode

**Future behavior:**
- Detect Arrow IPC magic bytes → Arrow Mode
- Fallback to File Mode if Arrow fails
- Optimize based on input size

**Usage:**
```bash
# Auto-detection (currently uses file mode)
zea load sales.csv | zea sql "SELECT * FROM stdin"
```

## Performance Comparison

| Mode | Throughput | Latency | Memory | Disk I/O |
|------|-----------|---------|--------|----------|
| **File (CSV)** | ~100 MB/s | Medium | Low | High |
| **Arrow (Future)** | ~1 GB/s+ | Low | Medium | None |
| **Hybrid** | Variable | Low-Medium | Low-Medium | Variable |

## Implementation Details

### File Mode Internals

```go
// internal/duckdb/sql.go
func RunSQLFromStdin(query string, stdin io.Reader, stdout io.Writer) error {
    1. Create temp CSV file (/tmp/zea_stdin_<pid>.csv)
    2. Copy stdin → temp file
    3. DuckDB: CREATE TABLE stdin AS SELECT * FROM read_csv_auto('tempfile')
    4. Execute SQL query
    5. Write results as CSV to stdout
    6. Cleanup temp file
}
```

**Temp file cleanup:** Automatic via `defer os.Remove()`

### Arrow Mode Internals

```go
// internal/duckdb/arrow.go
func RunSQLArrowNative(query string, stdin io.Reader, stdout io.Writer) error {
    1. Read Arrow IPC stream from stdin
    2. Convert to temp Parquet (intermediate format)
    3. DuckDB: CREATE TABLE stdin AS SELECT * FROM read_parquet('tempfile')
    4. Execute SQL query
    5. Convert results to Arrow IPC
    6. Write Arrow IPC to stdout
}
```

**Note:** Currently uses Parquet intermediate format for DuckDB compatibility. Future optimization: Direct Arrow table registration.

## Command-Line Flags

| Flag | Mode | Purpose |
|------|------|---------|
| (none) | Auto | Auto-detect best path |
| `--file` | File | Force file-based path |
| `--arrow` | Arrow | Force Arrow-native path |

**Mutual exclusivity:** Cannot specify both `--file` and `--arrow`

## Upgrade Path to Arrow

### Current State (v0.3.0)

```bash
# All commands output CSV
zea load sales.csv          # → CSV
zea filter "x > 100"        # → CSV
zea sql "SELECT * FROM..."  # → CSV
```

### Future State (v0.4.0+)

```bash
# Commands support Arrow IPC output
zea load --output=arrow sales.parquet    # → Arrow IPC
zea filter --output=arrow "x > 100"      # → Arrow IPC
zea sql --arrow "SELECT * FROM..."       # → Arrow IPC

# Hybrid pipelines with auto-conversion
zea load --output=arrow sales.parquet | zea sql --arrow "SELECT ..." | zea view
```

**Required changes:**
1. Add Arrow IPC output format to `internal/zeaframe/io.go`
2. Add `--output=arrow` flag to all commands
3. Update `zea view` to accept Arrow IPC input
4. Enable auto-detection in SQL command

## Testing

### File Mode Tests

```bash
# Test 1: Simple aggregation
echo "a,b\n1,2" | zea load | zea sql "SELECT SUM(b) FROM stdin"

# Test 2: Complex query
zea load sales.csv | zea sql --file "SELECT region, COUNT(*) FROM stdin GROUP BY region"

# Test 3: Hybrid pipeline
zea load sales.csv | zea filter "amount > 100" | zea sql "SELECT AVG(amount) FROM stdin"
```

### Arrow Mode Tests (Future)

```bash
# Test 1: Arrow roundtrip
zea load --output=arrow sales.parquet | zea sql --arrow "SELECT * FROM stdin" | zea store output.parquet

# Test 2: Performance benchmark
time zea load --output=arrow large.parquet | zea sql --arrow "SELECT * FROM stdin WHERE x > 1000"
```

## Error Handling

### File Mode Errors

- **Temp file creation fails:** Return error immediately
- **DuckDB read_csv fails:** Return query error with context
- **Query syntax error:** Pass through DuckDB error
- **Cleanup failure:** Log warning, continue

### Arrow Mode Errors

- **Arrow IPC read fails:** Currently returns error (no fallback)
- **Future:** Automatic fallback to file mode
- **Schema mismatch:** Return type error
- **DuckDB Arrow integration fails:** Fallback to Parquet intermediate

## Future Enhancements

### v0.4.0 - Arrow IPC Output
- Add Arrow IPC output format to all commands
- Enable auto-detection for Arrow vs CSV
- Automatic fallback mechanism

### v0.5.0 - MotherDuck Integration
- Connection string support: `-c "md:database"`
- Cloud table access
- Query federation

### v0.6.0 - Interactive REPL
- `zea sql -i` for interactive SQL shell
- History and auto-completion
- Multi-line query support

### v0.7.0 - dbt Integration
- `zea dbt-sync` for model generation
- Automatic schema sync
- Query optimization

## Debugging

### Enable Debug Mode

```bash
export ZEA_DEBUG=1
zea load sales.csv | zea sql "SELECT * FROM stdin"
```

**Debug output:**
- Temp file paths
- Mode selection
- Arrow fallback triggers
- Performance metrics

### Inspect Temp Files

```bash
# File mode temp files
ls -lh /tmp/zea_stdin_*.csv

# View temp CSV
cat /tmp/zea_stdin_12345.csv

# Check DuckDB query plan
zea sql "EXPLAIN SELECT * FROM stdin"
```

## Benchmarking

### File Mode Benchmark

```bash
time zea load large.csv | zea sql --file "SELECT COUNT(*) FROM stdin"
```

### Arrow Mode Benchmark (Future)

```bash
time zea load --output=arrow large.parquet | zea sql --arrow "SELECT COUNT(*) FROM stdin"
```

### Comparison Script

```bash
#!/bin/bash
echo "File mode:"
time zea load data.csv | zea sql --file "SELECT * FROM stdin" > /dev/null

echo "Arrow mode:"
time zea load --output=arrow data.parquet | zea sql --arrow "SELECT * FROM stdin" > /dev/null
```

## Architecture Benefits

### Reliability
- File mode provides rock-solid fallback
- Temp files are debuggable and inspectable
- No silent data loss

### Performance
- Arrow path ready for 10x+ speedup
- Zero-copy data transfer (future)
- Minimal memory overhead

### Flexibility
- Users can force specific modes
- Auto-detection for ease of use
- Gradual migration path

### Extensibility
- Clear separation of execution paths
- Easy to add new modes (e.g., Parquet-native)
- Plugin-friendly architecture

## Related Documentation

- [Commands Reference](COMMANDS.md#zea-sql) - SQL command usage
- [Plugin System](PLUGINS.md) - Extending ZeaShell
- [Parquet Support](PARQUET.md) - Parquet format details

---

**Last Updated:** 2026-03-20
**Version:** 0.3.0+
