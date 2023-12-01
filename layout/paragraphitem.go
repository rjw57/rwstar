package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

type ParagraphItemType int

const (
	ParagraphItemTypeBox     ParagraphItemType = 0
	ParagraphItemTypeGlue                      = iota
	ParagraphItemTypePenalty                   = iota
)

type ParagraphItemPenalty int

const (
	ParagraphItemPenaltyNever   ParagraphItemType = 1000
	ParagraphItemPenaltyNeutral                   = 0
	ParagraphItemPenaltyAlways                    = -1000
)

// ParagraphItem represents a layout item within a paragraph. Items may be boxes, glue or penalties.
//
// Boxes are literal horizontal collections of Cells to be rendered. Glue are Cells which are
// rendered representing the space between words. Penalties represent explicit line-breaking
// opportunity points.
type ParagraphItem struct {
	// Type represents the type of this layout item: box, glue or penalty.
	Type ParagraphItemType

	// Text is the content of this item when rendered on screen. Only applicable to boxes.
	Text string

	// Style is the appearance of this item when rendered on screen.
	Style tcell.Style

	// StartOffset is the lowest inclusive offset within the underlying paragraph represented by
	// this item.
	StartOffset int

	// EndOffset is the lowest offset within the underlying paragraph >= StartOffset which is not
	// represented by this item.
	EndOffset int

	// Penalty gives a penalty for breaking the line at this item. If the penalty is +ve, the line
	// will never be broken. If the penalty is 0, the line _may_ be broken. If the penalty is -ve
	// the line will always be broken.
	//
	// Only glue and penalties can break lines. Penalty is ignored for boxes.
	Penalty ParagraphItemPenalty
}

// CellCount is the *minimum* number of on-screen cells required to represent the item. Glue, in
// particular, may be rendered with more cells.
func (p *ParagraphItem) CellCount() int {
	return uniseg.StringWidth(p.Text)
}
