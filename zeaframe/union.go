package zeaframe

import (
	"fmt"
	"sync"
)

// LoadMultipleFiles loads multiple files in parallel and unions them into a single ZeaFrame
func LoadMultipleFiles(paths []string, opts *GlobOptions) (*ZeaFrame, error) {
	if opts == nil {
		opts = DefaultGlobOptions()
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no files to load")
	}

	// Single file - use existing path
	if len(paths) == 1 {
		return LoadAuto(paths[0])
	}

	// Infer schema from first few files
	schema, err := inferSchemaFromFiles(paths, 3)
	if err != nil {
		return nil, fmt.Errorf("schema inference failed: %w", err)
	}

	// Load files in parallel
	workers := opts.Parallel
	if workers <= 0 {
		workers = 8
	}

	type loadResult struct {
		zf  *ZeaFrame
		err error
		idx int
	}

	jobs := make(chan int, len(paths))
	results := make(chan loadResult, len(paths))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				zf, err := LoadAuto(paths[idx])
				results <- loadResult{zf: zf, err: err, idx: idx}
			}
		}()
	}

	// Queue jobs
	go func() {
		for i := range paths {
			jobs <- i
		}
		close(jobs)
	}()

	// Wait for workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	frames := make([]*ZeaFrame, len(paths))
	for result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", paths[result.idx], result.err)
		}
		frames[result.idx] = result.zf
	}

	// Union all frames
	return UnionFrames(frames, schema)
}

// inferSchemaFromFiles infers a unified schema from the first N files
func inferSchemaFromFiles(paths []string, sampleSize int) (map[string]ColumnType, error) {
	if sampleSize > len(paths) {
		sampleSize = len(paths)
	}

	schema := make(map[string]ColumnType)

	for i := 0; i < sampleSize; i++ {
		zf, err := LoadAuto(paths[i])
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", paths[i], err)
		}

		// Merge schema
		for _, col := range zf.Columns {
			existingType, exists := schema[col.Name]
			if !exists {
				schema[col.Name] = col.Type
			} else if existingType != col.Type {
				// Type mismatch - use most general type
				schema[col.Name] = promoteType(existingType, col.Type)
			}
		}
	}

	return schema, nil
}

// promoteType returns the most general type between two types
func promoteType(t1, t2 ColumnType) ColumnType {
	// Priority: StringType > Float64Type > Int64Type > BoolType
	priority := map[ColumnType]int{
		StringType:  4,
		Float64Type: 3,
		Int64Type:   2,
		BoolType:    1,
		MultiType:   5,
	}

	p1, p2 := priority[t1], priority[t2]
	if p1 > p2 {
		return t1
	}
	return t2
}

// UnionFrames combines multiple ZeaFrames into one, handling schema evolution
func UnionFrames(frames []*ZeaFrame, schema map[string]ColumnType) (*ZeaFrame, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames to union")
	}

	if len(frames) == 1 {
		return frames[0], nil
	}

	// Create result frame with unified schema
	result := NewZeaFrame()

	// Get column names in consistent order
	colNames := make([]string, 0, len(schema))
	for name := range schema {
		colNames = append(colNames, name)
	}
	// Sort for deterministic ordering
	// (Using simple iteration order for now - could sort alphabetically)

	// Initialize columns in result
	for name, colType := range schema {
		col := &Column{
			Name:  name,
			Type:  colType,
			Data:  make([]interface{}, 0),
			Nulls: make([]bool, 0),
		}
		result.Columns = append(result.Columns, col)
	}

	// Append data from each frame
	for _, frame := range frames {
		// Build column index for this frame
		frameColIdx := make(map[string]int)
		for i, col := range frame.Columns {
			frameColIdx[col.Name] = i
		}

		// Append rows
		for rowIdx := 0; rowIdx < frame.Rows; rowIdx++ {
			for _, col := range result.Columns {
				// Find column in source frame
				srcIdx, exists := frameColIdx[col.Name]

				if exists {
					// Copy value from source
					srcCol := frame.Columns[srcIdx]
					col.Data = append(col.Data, srcCol.Data[rowIdx])
					col.Nulls = append(col.Nulls, srcCol.Nulls[rowIdx])
				} else {
					// Column doesn't exist in this frame - use NULL
					col.Data = append(col.Data, "")
					col.Nulls = append(col.Nulls, true)
				}
			}
			result.Rows++
		}
	}

	return result, nil
}

// PreviewSchema shows the inferred schema without loading data
func PreviewSchema(paths []string) (map[string]ColumnType, error) {
	return inferSchemaFromFiles(paths, 3)
}
