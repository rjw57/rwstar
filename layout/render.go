package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/rjw57/rwstar/document"
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

var (
	StyleNormal = tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorLightGray)
	StyleGlue   = StyleNormal.Foreground(tcell.ColorDarkGray)
	StyleMarkup = StyleNormal.Foreground(tcell.ColorDarkCyan)
)

// ParagraphItem represents a layout item within a paragraph. Items may be boxes, glue or penalties.
//
// Boxes are literal horizontal collections of Cells to be rendered. Glue are Cells which are
// rendered representing the space between words. Penalties represent explicit line-breaking
// opportunity points.
type ParagraphItem struct {
	// Type represents the type of this layout item: box, glue or penalty.
	Type ParagraphItemType

	// Text is the on-screen content of this item.
	Text string

	// Style is the on-screen appearance of this item.
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

func (p *ParagraphItem) CellCount() int {
	return uniseg.StringWidth(p.Text)
}

func appendTextParagraphItems(items []ParagraphItem, text string, startOffset int) []ParagraphItem {
	state := -1
	var segment string

	for len(text) > 0 {
		segment, text, _, state = uniseg.FirstLineSegmentInString(text, state)
		items = appendLineSegmentParagraphItems(items, segment, startOffset)
		startOffset += len(segment)

		// If the segment ends with a forced line break, add a penalty.
		if uniseg.HasTrailingLineBreakInString(segment) {
			items = append(items, ParagraphItem{
				Type:        ParagraphItemTypePenalty,
				StartOffset: startOffset,
				EndOffset:   startOffset,
				Penalty:     ParagraphItemPenaltyAlways,
			})
		}
	}

	return items
}

func appendLineSegmentParagraphItems(items []ParagraphItem, text string, startOffset int) []ParagraphItem {
	state := -1
	var word string

	for len(text) > 0 {
		word, text, state = uniseg.FirstWordInString(text, state)

		// Multiple spaces become multiple glues
		for len(word) > 0 && word[0] == ' ' {
			item := ParagraphItem{
				Type:        ParagraphItemTypeGlue,
				Text:        word[:1],
				Style:       StyleGlue,
				StartOffset: startOffset,
				EndOffset:   startOffset + 1,
			}
			items = append(items, item)
			word = word[1:]
			startOffset++
		}

		if len(word) > 0 {
			item := ParagraphItem{
				Type:        ParagraphItemTypeBox,
				Text:        word,
				Style:       StyleNormal,
				StartOffset: startOffset,
				EndOffset:   startOffset + len(word),
			}
			items = append(items, item)
			startOffset += len(word)
		}
	}

	return items
}

func (l *Layout) renderParagraphLines(p *document.Paragraph) Lines {
	var lines Lines
	var items []ParagraphItem

	text := p.String()
	items = appendTextParagraphItems(items, text, 0)

	// add forced line break
	items = append(items, []ParagraphItem{{
		Type:        ParagraphItemTypeBox,
		Text:        "¶",
		Style:       StyleMarkup,
		StartOffset: len(text) + 1,
		EndOffset:   len(text) + 1,
	}, {
		Type:        ParagraphItemTypePenalty,
		StartOffset: len(text) + 1,
		EndOffset:   len(text) + 1,
		Penalty:     ParagraphItemPenaltyAlways,
	}}...)

	if items[len(items)-1].Penalty != ParagraphItemPenaltyAlways || items[len(items)-1].Type != ParagraphItemTypePenalty {
		panic("Paragraph items do not end in forced break.")
	}

	// array giving running width of boxes up to but not including item at that index
	runningWidths := make([]int, 1, len(items)+1)
	for _, item := range items {
		runningWidths = append(runningWidths, runningWidths[len(runningWidths)-1]+item.CellCount())
	}

	lineStartIdx := 0
	lineBreakIdx := -1
	for itemIdx, item := range items {
		// We can never break on boxes
		if item.Type == ParagraphItemTypeBox {
			continue
		}

		// Compute width of a line from current start breaking here.
		w := runningWidths[itemIdx] - runningWidths[lineStartIdx]

		// If this is feasible, record it.
		if w <= l.screenWidth {
			lineBreakIdx = itemIdx
		}

		// If not feasible or is forced, record line.
		if w > l.screenWidth || item.Penalty < 0 {
			lines = append(lines, items[lineStartIdx:lineBreakIdx])
			lineStartIdx = lineBreakIdx + 1
		}
	}

	return lines
}
