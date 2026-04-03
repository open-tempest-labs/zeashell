package zeaframe

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
)

// FromArrow converts Arrow schema + records directly into a ZeaFrame.
// No IPC serialisation occurs — the Arrow column buffers are read in place.
// Callers retain ownership of the records; this function does not release them.
func FromArrow(schema *arrow.Schema, records []arrow.Record) (*ZeaFrame, error) {
	zf := NewZeaFrame()
	if schema == nil || len(records) == 0 {
		return zf, nil
	}

	numCols := schema.NumFields()
	colData := make([][]interface{}, numCols)
	colNulls := make([][]bool, numCols)
	for i := range colData {
		colData[i] = make([]interface{}, 0)
		colNulls[i] = make([]bool, 0)
	}

	totalRows := 0
	for _, rec := range records {
		n := int(rec.NumRows())
		totalRows += n
		for colIdx := 0; colIdx < numCols; colIdx++ {
			col := rec.Column(colIdx)
			ct := arrowTypeToColumnType(schema.Field(colIdx).Type)
			for rowIdx := 0; rowIdx < n; rowIdx++ {
				isNull := col.IsNull(rowIdx)
				colNulls[colIdx] = append(colNulls[colIdx], isNull)
				if isNull {
					colData[colIdx] = append(colData[colIdx], getZeroValue(ct))
				} else {
					colData[colIdx] = append(colData[colIdx], arrowValueAt(col, rowIdx))
				}
			}
		}
	}

	for colIdx := 0; colIdx < numCols; colIdx++ {
		field := schema.Field(colIdx)
		ct := arrowTypeToColumnType(field.Type)
		if err := zf.AddColumn(field.Name, ct, colData[colIdx]); err != nil {
			return nil, err
		}
		zf.Columns[colIdx].Nulls = colNulls[colIdx]
	}
	zf.Rows = totalRows
	return zf, nil
}

func arrowTypeToColumnType(dt arrow.DataType) ColumnType {
	switch dt.ID() {
	case arrow.INT8, arrow.INT16, arrow.INT32, arrow.INT64,
		arrow.UINT8, arrow.UINT16, arrow.UINT32, arrow.UINT64:
		return Int64Type
	case arrow.FLOAT32, arrow.FLOAT64:
		return Float64Type
	case arrow.BOOL:
		return BoolType
	default:
		return StringType
	}
}

func arrowValueAt(col arrow.Array, i int) interface{} {
	switch c := col.(type) {
	case *array.Int8:
		return int64(c.Value(i))
	case *array.Int16:
		return int64(c.Value(i))
	case *array.Int32:
		return int64(c.Value(i))
	case *array.Int64:
		return c.Value(i)
	case *array.Uint8:
		return int64(c.Value(i))
	case *array.Uint16:
		return int64(c.Value(i))
	case *array.Uint32:
		return int64(c.Value(i))
	case *array.Uint64:
		return int64(c.Value(i))
	case *array.Float32:
		return float64(c.Value(i))
	case *array.Float64:
		return c.Value(i)
	case *array.Boolean:
		return c.Value(i)
	case *array.String:
		return c.Value(i)
	case *array.LargeString:
		return c.Value(i)
	default:
		if v := col.GetOneForMarshal(i); v != nil {
			return fmt.Sprintf("%v", v)
		}
		return nil
	}
}
