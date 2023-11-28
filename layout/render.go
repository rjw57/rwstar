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
	StyleGlue   = StyleNormal.Foreground(tcell.ColorGrey)
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

	// Cells contains the visual representation of this item if this item is a box.
	Cells []Cell
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
				StartOffset: startOffset,
				EndOffset:   startOffset + 1,
			}
			item.Cells = []Cell{{
				Mainc:       tcell.RuneBullet,
				Style:       StyleGlue,
				StartOffset: item.StartOffset,
				EndOffset:   item.EndOffset,
			}}
			items = append(items, item)
			word = word[1:]
			startOffset++
		}

		if len(word) > 0 {
			item := ParagraphItem{
				Type:        ParagraphItemTypeBox,
				StartOffset: startOffset,
				EndOffset:   startOffset + len(word),
			}
			item.Cells = appendWordCells(item.Cells, word, startOffset)
			items = append(items, item)
			startOffset += len(word)
		}
	}

	return items
}

func appendWordCells(cells []Cell, text string, startOffset int) []Cell {
	state := -1
	var cluster string

	for len(text) > 0 {
		var width int
		cluster, text, width, state = uniseg.FirstGraphemeClusterInString(text, state)

		clusterRunes := []rune(cluster)
		cell := Cell{
			StartOffset: startOffset,
			EndOffset:   startOffset + len(cluster),
			Mainc:       clusterRunes[0],
			Combc:       clusterRunes[1:],
			Style:       StyleNormal,
		}
		cells = append(cells, cell)

		for ; width > 1; width-- {
			cells = append(cells, Cell{})
		}

		startOffset += len(cluster)
	}

	return cells
}

func (l *Layout) renderParagraphLines(p *document.Paragraph) Lines {
	var lines Lines
	var items []ParagraphItem

	text := p.String()
	items = appendTextParagraphItems(items, text, 0)

	// add forced line break
	items = append(items, ParagraphItem{
		Type:        ParagraphItemTypePenalty,
		StartOffset: len(text),
		EndOffset:   len(text),
		Penalty:     ParagraphItemPenaltyAlways,
	})

	// array giving running width of boxes up to but not including item at that index
	runningWidths := make([]int, 1, len(items)+1)
	for _, item := range items {
		runningWidths = append(runningWidths, runningWidths[len(runningWidths)-1]+len(item.Cells))
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
			startItem := items[lineStartIdx]
			line := Line{StartOffset: startItem.StartOffset, EndOffset: item.StartOffset}
			for _, lineItem := range items[lineStartIdx:lineBreakIdx] {
				line.Cells = append(line.Cells, lineItem.Cells...)
			}
			lines = append(lines, line)
			lineStartIdx = lineBreakIdx + 1
		}
	}

	return lines
}
