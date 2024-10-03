package column

import (
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
)

type FloatHandler struct {
	field arrow.Field
	items []float64
	valid []bool
	index int

	column     *arrow.Column
	chunkIndex int
	pos        int
}

func NewFloatHandler(name string, index int, nullable bool) *FloatHandler {
	field := arrow.Field{
		Name:     name,
		Type:     arrow.PrimitiveTypes.Float64,
		Nullable: nullable,
	}

	return &FloatHandler{
		field: field,
		items: make([]float64, 0),
		valid: make([]bool, 0),
		index: index,
	}
}

func (h *FloatHandler) Add(v any) error {
	if v == nil {
		h.items = append(h.items, 0)
		h.valid = append(h.valid, false)
		return nil
	}
	h.valid = append(h.valid, true)
	switch val := v.(type) {
	case float32:
		h.items = append(h.items, float64(val))
	case float64:
		h.items = append(h.items, val)
	case *float64:
		if val == nil {
			h.items = append(h.items, 0)
			h.valid[len(h.valid)-1] = false
		} else {
			h.items = append(h.items, *val)
		}
	default:
		return fmt.Errorf("cannot convert %v of type %T to float64", v, v)
	}
	return nil
}

func (h *FloatHandler) Build(builder *array.RecordBuilder) {
	builder.Field(h.index).(*array.Float64Builder).AppendValues(h.items, h.valid)
}

func (h *FloatHandler) GetScanType() any {
	return new(float64)
}

func (h *FloatHandler) SetColumn(column *arrow.Column) {
	h.column = column
	h.Reset()
}

func (h *FloatHandler) Next() bool {
	h.pos++
	chunks := h.column.Data().Chunks()
	// 如果 position 超了，但是 chunkIndex 还有，改到下一个
	if h.pos >= chunks[h.chunkIndex].Len() {
		if h.chunkIndex < len(chunks)-1 {
			h.chunkIndex++
			h.pos = 0
			return true
		} else {
			return false
		}
	}
	return h.pos < chunks[h.chunkIndex].Len()
}

func (h *FloatHandler) Value() any {
	chunks := h.column.Data().Chunks()
	chunk := chunks[h.chunkIndex].(*array.Float64)
	if chunk.IsNull(h.pos) {
		return nil
	}
	return chunk.Value(h.pos)
}

func (h *FloatHandler) Reset() {
	h.chunkIndex = 0
	h.pos = -1
}

func (h *FloatHandler) GetArrowField() arrow.Field {
	return h.field
}
