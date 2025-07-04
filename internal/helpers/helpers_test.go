package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplication_BuildValuePlaceholders(t *testing.T) {
	tests := []struct {
		name    string
		numCols int
		numRows int
		want    string
	}{
		{
			name:    "numCols = numRows = 3",
			numCols: 3,
			numRows: 3,
			want:    "($1, $2, $3), ($4, $5, $6), ($7, $8, $9)",
		},
		{
			name:    "numCols = numRows = 0",
			numCols: 0,
			numRows: 0,
			want:    "",
		},
		{
			name:    "numCols > numRows",
			numCols: 3,
			numRows: 5,
			want:    "($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15)",
		},
		{
			name:    "numCols < numRows",
			numCols: 5,
			numRows: 3,
			want:    "($1, $2, $3, $4, $5), ($6, $7, $8, $9, $10), ($11, $12, $13, $14, $15)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildValuePlaceholders(tt.numCols, tt.numRows)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestApplication_SplitIntoBatches(t *testing.T) {
	tests := []struct {
		name      string
		data      []int
		batchSize int
		want      [][]int
	}{
		{
			name:      "Test 1",
			data:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			batchSize: 3,
			want:      [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := SplitIntoBatches(tt.data, tt.batchSize)
			assert.Equal(t, tt.want, batches)
		})
	}
}
