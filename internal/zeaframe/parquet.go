package zeaframe

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/compress"
	"github.com/apache/arrow/go/v18/parquet/file"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
)

// FromParquet creates a ZeaFrame from a Parquet file
func FromParquet(path string) (*ZeaFrame, error) {
	// Open the parquet file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	pqFile, err := file.NewParquetReader(f, file.WithReadProps(parquet.NewReaderProperties(memory.NewGoAllocator())))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file reader: %w", err)
	}
	defer pqFile.Close()

	// Create Arrow file reader
	mem := memory.NewGoAllocator()
	fileReader, err := pqarrow.NewFileReader(pqFile, pqarrow.ArrowReadProperties{}, mem)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	// Read entire table with context
	ctx := context.Background()
	table, err := fileReader.ReadTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	// Convert Arrow table to ZeaFrame
	return arrowTableToZeaFrame(table)
}

// WriteParquet writes the ZeaFrame to a Parquet file
func (zf *ZeaFrame) WriteParquet(path string) error {
	return zf.WriteParquetWithOptions(path, compress.Codecs.Snappy, 128*1024*1024) // 128MB row groups
}

// WriteParquetWithOptions writes the ZeaFrame to a Parquet file with custom options
func (zf *ZeaFrame) WriteParquetWithOptions(path string, compression compress.Compression, rowGroupSize int64) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer file.Close()

	// Convert ZeaFrame to Arrow table
	table, err := zf.toArrowTable()
	if err != nil {
		return fmt.Errorf("failed to convert to arrow table: %w", err)
	}
	defer table.Release()

	// Create Arrow file writer with properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(compression),
		parquet.WithDictionaryDefault(true),
	)

	arrowProps := pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema())

	writer, err := pqarrow.NewFileWriter(
		table.Schema(),
		file,
		props,
		arrowProps,
	)
	if err != nil {
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}
	defer writer.Close()

	// Write the table
	if err := writer.WriteTable(table, rowGroupSize); err != nil {
		return fmt.Errorf("failed to write parquet table: %w", err)
	}

	return nil
}

// arrowTableToZeaFrame converts an Arrow table to a ZeaFrame
func arrowTableToZeaFrame(table arrow.Table) (*ZeaFrame, error) {
	zf := NewZeaFrame()
	schema := table.Schema()

	numRows := int(table.NumRows())
	zf.Rows = numRows

	// Process each column
	for i := 0; i < int(table.NumCols()); i++ {
		col := *table.Column(i)
		field := schema.Field(i)

		// Convert Arrow column to ZeaFrame column
		zeaCol, err := arrowColumnToZeaColumn(field.Name, col, numRows)
		if err != nil {
			return nil, fmt.Errorf("failed to convert column %s: %w", field.Name, err)
		}

		zf.Columns = append(zf.Columns, zeaCol)
	}

	return zf, nil
}

// arrowColumnToZeaColumn converts an Arrow column to a ZeaFrame column
func arrowColumnToZeaColumn(name string, col arrow.Column, numRows int) (*Column, error) {
	data := make([]interface{}, numRows)
	nulls := make([]bool, numRows)

	// Process each chunk
	rowOffset := 0
	for _, chunk := range col.Data().Chunks() {
		chunkLen := chunk.Len()

		switch arr := chunk.(type) {
		case *array.String:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = ""
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = arr.Value(i)
				}
			}

		case *array.Int64:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = int64(0)
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = arr.Value(i)
				}
			}

		case *array.Int32:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = int64(0)
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = int64(arr.Value(i))
				}
			}

		case *array.Float64:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = 0.0
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = arr.Value(i)
				}
			}

		case *array.Float32:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = 0.0
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = float64(arr.Value(i))
				}
			}

		case *array.Boolean:
			for i := 0; i < chunkLen; i++ {
				if arr.IsNull(i) {
					data[rowOffset+i] = false
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = arr.Value(i)
				}
			}

		default:
			// Fallback: convert to string
			for i := 0; i < chunkLen; i++ {
				if chunk.IsNull(i) {
					data[rowOffset+i] = ""
					nulls[rowOffset+i] = true
				} else {
					data[rowOffset+i] = chunk.ValueStr(i)
				}
			}
		}

		rowOffset += chunkLen
	}

	// Infer ZeaFrame column type
	colType := inferTypeFromArrow(col.DataType())

	return &Column{
		Name:  name,
		Type:  colType,
		Data:  data,
		Nulls: nulls,
	}, nil
}

// toArrowTable converts a ZeaFrame to an Arrow table
func (zf *ZeaFrame) toArrowTable() (arrow.Table, error) {
	mem := memory.NewGoAllocator()

	// Build Arrow schema
	fields := make([]arrow.Field, len(zf.Columns))
	for i, col := range zf.Columns {
		fields[i] = arrow.Field{
			Name: col.Name,
			Type: zeaTypeToArrowType(col.Type),
		}
	}
	schema := arrow.NewSchema(fields, nil)

	// Build record batches
	builders := make([]array.Builder, len(zf.Columns))
	for i, col := range zf.Columns {
		builders[i] = createArrowBuilder(mem, col.Type)
	}

	// Populate builders
	for rowIdx := 0; rowIdx < zf.Rows; rowIdx++ {
		for colIdx, col := range zf.Columns {
			if col.Nulls[rowIdx] {
				builders[colIdx].AppendNull()
			} else {
				appendValueToBuilder(builders[colIdx], col.Data[rowIdx], col.Type)
			}
		}
	}

	// Create arrays and build columns
	cols := make([]arrow.Column, len(zf.Columns))
	for i, builder := range builders {
		arr := builder.NewArray()
		defer arr.Release()

		chunked := arrow.NewChunked(fields[i].Type, []arrow.Array{arr})
		cols[i] = *arrow.NewColumn(fields[i], chunked)
	}

	// Create table from columns and row count
	table := array.NewTable(schema, cols, int64(zf.Rows))
	return table, nil
}

// inferTypeFromArrow infers a ZeaFrame type from an Arrow type
func inferTypeFromArrow(arrowType arrow.DataType) ColumnType {
	switch arrowType.ID() {
	case arrow.STRING, arrow.LARGE_STRING:
		return StringType
	case arrow.INT64, arrow.INT32, arrow.INT16, arrow.INT8:
		return Int64Type
	case arrow.FLOAT64, arrow.FLOAT32:
		return Float64Type
	case arrow.BOOL:
		return BoolType
	default:
		return StringType
	}
}

// zeaTypeToArrowType converts a ZeaFrame type to an Arrow type
func zeaTypeToArrowType(zeaType ColumnType) arrow.DataType {
	switch zeaType {
	case StringType:
		return arrow.BinaryTypes.String
	case Int64Type:
		return arrow.PrimitiveTypes.Int64
	case Float64Type:
		return arrow.PrimitiveTypes.Float64
	case BoolType:
		return arrow.FixedWidthTypes.Boolean
	default:
		return arrow.BinaryTypes.String
	}
}

// createArrowBuilder creates an appropriate Arrow array builder for a ZeaFrame type
func createArrowBuilder(mem memory.Allocator, zeaType ColumnType) array.Builder {
	switch zeaType {
	case StringType:
		return array.NewStringBuilder(mem)
	case Int64Type:
		return array.NewInt64Builder(mem)
	case Float64Type:
		return array.NewFloat64Builder(mem)
	case BoolType:
		return array.NewBooleanBuilder(mem)
	default:
		return array.NewStringBuilder(mem)
	}
}

// appendValueToBuilder appends a value to an Arrow array builder
func appendValueToBuilder(builder array.Builder, value interface{}, colType ColumnType) {
	switch b := builder.(type) {
	case *array.StringBuilder:
		b.Append(fmt.Sprintf("%v", value))
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			b.Append(v)
		} else {
			b.Append(0)
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			b.Append(v)
		} else {
			b.Append(0.0)
		}
	case *array.BooleanBuilder:
		if v, ok := value.(bool); ok {
			b.Append(v)
		} else {
			b.Append(false)
		}
	}
}
