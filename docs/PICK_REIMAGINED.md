# PICK Reimagined: Multi-Valued Data for the Modern Age

This document explores the history of PICK Operating System and related systems, their revolutionary approach to data management, and how their philosophy inspires ZeaShell's design for modern data processing.

## The PICK Legacy

### History

The **PICK Operating System** was developed in 1965 by Dick Pick and Don Nelson at TRW for the U.S. Army. Initially created as a database system for managing helicopter parts inventory, PICK evolved into a complete operating system with an integrated database, programming language, and query tools.

**Key Historical Milestones:**

- **1965**: Original development at TRW as "Generalized Information Retrieval Language System" (GIRLS)
- **1973**: Renamed to PICK after Dick Pick
- **1970s-1980s**: Licensed to multiple vendors, creating diverse implementations
- **1980s-1990s**: Peak popularity with millions of users worldwide
- **2000s**: Decline due to proprietary nature, but systems still in production today

### PICK Implementations and Related Systems

PICK's success led to numerous implementations and variants:

**Major Implementations:**

- **Ultimate** (Ultimate Corp) - One of the earliest and most popular implementations
- **Prime INFORMATION** - Sold by Prime Computer
- **Reality** - From Microdata/McDonnell Douglas
- **UniData/UniVerse** (Rocket Software) - Modern commercial implementations still in use
- **jBASE** - Java-based PICK implementation
- **OpenQM** - Open-source PICK implementation
- **D3** - Windows-based PICK system

**Related Systems:**

- **Universe/Pick** - IBM's implementation after acquiring Ardent
- **Revelation** - PC-based PICK system
- **mvBase** - Another PC implementation
- **PickBlue** - Blue Cross/Blue Shield's variant

These systems powered critical business applications across industries: healthcare, manufacturing, distribution, retail, and finance. Many legacy PICK systems remain in production decades later, testament to their robustness and productivity.

## The PICK Philosophy

### Multi-Valued Databases

PICK's revolutionary concept was the **multi-valued database** - a hierarchical data model that differed fundamentally from relational databases:

**Multi-Valued Fields:**
- Fields could contain **multiple values** (arrays)
- Values could themselves be **multi-valued** (nested arrays)
- Sub-values created hierarchical structures
- No need for separate tables or foreign keys

**Example PICK Record:**
```
Customer ID: 12345
Name: Acme Corp
Orders: 1001^1002^1003
Order Dates: 2023-01-15^2023-02-20^2023-03-10
Order Amounts: 500^750^1200
```

Each customer record contained all their orders inline, with correlated multi-values (order 1001 was on 2023-01-15 for $500).

**Benefits:**
- **Natural data modeling** - Business entities as complete records
- **No joins required** - Related data stored together
- **Fast queries** - No table scans or index lookups across tables
- **Simple understanding** - Business users could grasp the model

### Data Dictionaries

PICK integrated **data dictionaries** directly into the database:

**Dictionary Attributes:**
- Column definitions with business names
- Computed/derived fields (virtual columns)
- Data validation rules
- Formatting specifications
- Correlative formulas

**Power of Dictionaries:**
- Defined once, used everywhere
- Business logic in the dictionary, not application code
- Changed dictionary → changed all queries/reports automatically
- Non-programmers could create new "columns" via correlatives

**Example:**
```
DICT CUSTOMERS
TOTAL.ORDERS
  Correlative: SUM(ORDER.AMOUNTS)
  Format: MD2
  Title: "Total Orders"
```

This created a virtual column `TOTAL.ORDERS` summing the multi-valued `ORDER.AMOUNTS` field, formatted as money with 2 decimals.

### Integrated Query and Programming

PICK bundled powerful tools in one environment:

**ENGLISH/ACCESS Query Language:**
- Natural language-like syntax
- Business users could write queries without SQL knowledge
- Example: `LIST CUSTOMERS WITH TOTAL.ORDERS > 10000 BY STATE`

**PICK BASIC:**
- Integrated programming language
- Direct database access without middleware
- Built-in functions for multi-valued field manipulation
- Proc language for simple automation

**TCL (Terminal Control Language):**
- Command-line interface combining OS and database operations
- Everything accessible from the prompt: files, programs, queries
- Scriptable and interactive

**Spreadsheet Integration:**
- Many PICK systems included spreadsheet capabilities
- Pull data directly from database into cells
- Formulas could reference database fields
- Instant reporting without export/import

### The SMB Advantage

PICK excelled in the **Small and Medium Business** market:

**Productivity Benefits:**

1. **All-in-One System**
   - OS, database, programming, query tools in one package
   - No separate database server, application server, web server
   - Reduced complexity and cost

2. **Rapid Development**
   - Data dictionaries eliminated repetitive code
   - Multi-valued fields avoided complex schemas
   - PICK BASIC was concise and business-focused
   - Changes propagated automatically via dictionaries

3. **Business User Empowerment**
   - ENGLISH queries accessible to non-programmers
   - Data dictionaries let users define new views
   - Ad-hoc reporting without IT bottleneck
   - Spreadsheet integration for familiar interface

4. **Vertical Market Applications**
   - Vendors built industry-specific applications (healthcare, manufacturing, distribution)
   - Tight integration with PICK's data model
   - Customizable via dictionaries without changing code
   - Lower total cost of ownership

5. **Developer Productivity**
   - No impedance mismatch (database model = application model)
   - No ORM complexity
   - Fewer lines of code for business logic
   - Rapid prototyping and iteration

**Real-World Impact:**

A PICK developer could build a complete business application in weeks that would take months in traditional environments. Small businesses could afford custom software. Vertical market ISVs could serve hundreds of customers with small development teams.

## The PICK Decline

Despite its power, PICK declined in the 2000s:

**Challenges:**
- **Proprietary ecosystem** - Vendor lock-in, high costs
- **Limited interoperability** - Difficult integration with other systems
- **Smaller developer pool** - Fewer trained developers over time
- **Relational database dominance** - SQL became the standard
- **Internet era** - Web-centric architectures favored different models
- **Enterprise shift** - Large enterprises standardized on Oracle, SQL Server, etc.

However, PICK's core ideas were valuable. The decline was due to execution and ecosystem, not fundamental design flaws.

## ZeaShell: PICK Concepts for Modern Data

ZeaShell brings PICK's most powerful concepts to modern data processing with today's formats and tools.

### Multi-Valued Data: JSON and XML

**PICK Concept:** Multi-valued fields with nested structures

**ZeaShell Implementation:** Native support for nested JSON and XML

```bash
# JSON with nested arrays (multi-valued fields)
zea load customers.json | zea filter "orders[0] > 1000"

# Automatic "ANY" semantics like PICK
zea load data.json | zea filter "orders > 1000"
# Matches if ANY order > 1000 (PICK-style implicit ANY)
```

**Why This Works:**
- JSON arrays are the modern equivalent of PICK multi-values
- Nested objects mirror PICK sub-values
- Path-based column names (`address.city`) similar to PICK dictionary paths
- No need to flatten into relational tables

### Data Dictionary Concepts: Path-Based Columns

**PICK Concept:** Data dictionaries with computed fields and business names

**ZeaShell Implementation:** Automatic flattening to path-based column names

```bash
# PICK: Define computed field in dictionary
# ZeaShell: Nested paths are automatic columns
zea load data.json | zea select customer,address.city,address.state
```

**Benefits:**
- Nested structures become queryable "columns" automatically
- No manual schema definition required
- Path names are self-documenting
- Business-friendly dotted notation

### Integrated Query: Pipeable Commands

**PICK Concept:** ENGLISH query language integrated with OS

**ZeaShell Implementation:** Unix pipe semantics with SQL-like filters

```bash
# Natural query flow
zea load customers.csv \
  | zea filter "total_orders > 10000" \
  | zea group state --sum=total_orders \
  | zea sort total_orders_sum:desc
```

**Advantages:**
- Composable operations (Unix philosophy)
- Readable, business-friendly syntax
- Streaming for memory efficiency
- Mix with other Unix tools

### Interactive Exploration: TUI Viewer

**PICK Concept:** Interactive terminal environment with immediate feedback

**ZeaShell Implementation:** `zea view` interactive TUI

```bash
zea view customers.csv
# Interactive: sort, filter, graph, export with keyboard shortcuts
```

**PICK-Inspired Features:**
- Immediate visual feedback
- Ad-hoc exploration without writing queries
- Business user friendly
- Export filtered views instantly

### Developer Productivity: Single Binary, Multiple Formats

**PICK Concept:** All-in-one integrated environment

**ZeaShell Implementation:** Single binary supporting all modern formats

```bash
# No separate tools for different formats
zea load data.csv | zea store data.parquet    # CSV → Parquet
zea load data.json | zea filter "x > 10"       # JSON query
zea load topology.xml | zea group service.role # XML processing
```

**Benefits:**
- One tool, many formats
- No format-specific complexity
- Seamless conversions
- Reduced learning curve

### Schema Evolution: Automatic Handling

**PICK Concept:** Flexible schema, add fields without disruption

**ZeaShell Implementation:** Automatic schema evolution for multi-file loading

```bash
# Files with different schemas union automatically
zea load "sales/*.csv"
# Missing columns → NULLs (like PICK's missing attributes)
```

**Advantages:**
- No schema migrations required
- Old and new data coexist
- Gradual evolution
- Resilient to schema changes

## What PICK Got Right

ZeaShell preserves these core PICK strengths:

1. **Data Model Matches Business Reality**
   - Hierarchical data (customers with orders) is natural
   - No artificial normalization
   - Multi-valued fields (JSON arrays) supported natively

2. **Developer Productivity**
   - Fewer abstractions between data and code
   - Composable operations (pipes vs. complex SQL)
   - Rapid iteration and exploration

3. **Business User Accessibility**
   - Interactive exploration (TUI viewer)
   - Readable filter syntax
   - Self-documenting path-based columns

4. **Integrated Environment**
   - Single tool for all formats
   - No middleware complexity
   - Seamless format conversions

5. **Practical Scale**
   - SMB to petabyte-scale data lakes
   - Streaming for memory efficiency
   - Parallel loading for performance

## What ZeaShell Improves

ZeaShell modernizes PICK's concepts:

1. **Open Standards**
   - CSV, JSON, XML, Parquet (not proprietary)
   - Unix pipes (universal composition)
   - Standard Go libraries (no vendor lock-in)

2. **Cloud-Native**
   - Works with S3, Azure, GCS via mounts
   - Glob patterns for data lakes
   - Parallel processing for scale

3. **Modern Formats**
   - Parquet for columnar efficiency
   - JSON for hierarchical data
   - XML for structured documents
   - Seamless interoperability

4. **Ecosystem Integration**
   - Pipes to Unix tools
   - HTTP/HTTPS loading
   - Standard file formats
   - No walled garden

5. **Performance at Scale**
   - Columnar storage (Parquet)
   - Parallel multi-file loading
   - Streaming I/O
   - Works from KB to TB

## The Modern PICK Philosophy

**PICK's Lesson:** The best database is the one that matches how businesses think about their data.

**ZeaShell's Approach:** Support modern hierarchical formats (JSON, XML) natively while providing powerful relational operations (filter, group, join) through Unix pipe semantics.

**Result:**
- Business users can explore data interactively (TUI viewer)
- Developers can build pipelines quickly (composable commands)
- Data engineers can process at scale (partitioned data, Parquet)
- Everyone works with standard formats (no vendor lock-in)

## Lessons for Modern Data

What can today's data engineers learn from PICK?

1. **Hierarchical Data is Natural**
   - Don't force everything into relational tables
   - JSON/XML/Parquet support nested structures
   - Query into arrays without flattening

2. **Integration Reduces Friction**
   - One tool for many formats
   - Composable operations
   - Minimal abstractions

3. **Interactive Exploration Matters**
   - Ad-hoc queries without writing code
   - Visual feedback (charts, tables)
   - Lower barrier to entry

4. **Schema Flexibility is Valuable**
   - Schemas evolve over time
   - Handle missing fields gracefully
   - Don't require upfront perfection

5. **Productivity over Purity**
   - Practical solutions over theoretical elegance
   - Rapid iteration over perfect design
   - Get results, then optimize

## Conclusion

PICK Operating System was ahead of its time in recognizing that:
- **Hierarchical data** models many business domains naturally
- **Integrated tools** increase productivity
- **Business user access** to data is valuable
- **Flexible schemas** adapt to changing requirements

ZeaShell brings these insights to modern data processing:
- Native support for hierarchical formats (JSON, XML)
- Integrated commands for all operations
- Interactive viewer for exploration
- Automatic schema evolution

The result is a tool that embodies PICK's productivity philosophy while embracing modern standards, cloud-native scale, and open-source collaboration.

**PICK's legacy lives on - not in its proprietary implementations, but in its timeless insights about how humans work with data.**

---

## Further Reading

### PICK History
- *The PICK Operating System* by Jonathan E. Sisk
- PICK Systems Wikipedia: https://en.wikipedia.org/wiki/Pick_operating_system
- International Spectrum (PICK community magazine archives)

### PICK Concepts
- Multi-valued databases and their advantages
- Data dictionary design patterns
- PICK BASIC programming techniques

### Modern Equivalents
- JSON for hierarchical data
- Document databases (MongoDB, CouchDB)
- Columnar formats (Parquet, Arrow)
- DataFrame libraries (Pandas, Polars)

### ZeaShell Documentation
- [NESTED_QUERIES.md](NESTED_QUERIES.md) - Multi-valued field queries
- [JSON_SUPPORT.md](JSON_SUPPORT.md) - Hierarchical data processing
- [COMMANDS.md](COMMANDS.md) - Integrated query commands
- [VIEWER.md](VIEWER.md) - Interactive exploration

**ZeaShell: Modern data processing inspired by timeless ideas** 🗄️✨
