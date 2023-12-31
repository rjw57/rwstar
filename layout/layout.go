package layout

import (
	"errors"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/rjw57/rwstar/document"
)

const cacheChunkSizeLog2 = 5
const cacheChunkSize = 1 << cacheChunkSizeLog2

var (
	ErrPointIsFromDifferentDocument = errors.New("Point is from different document")
	ErrPointNotFound                = errors.New("Point not found")
)

var (
	StyleNormal = tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorLightGray)
	StyleMarkup = StyleNormal.Foreground(tcell.ColorDarkCyan)
)

type Cell struct {
	// Mainc represents the main rune for this cell. If 0, the cell is the rightmost companion to a
	// wide character and should not be rendered.
	Mainc rune

	// Combc holds any combining runes for the cell.
	Combc []rune

	// Style describes how to render the cell.
	Style tcell.Style

	// StartOffset gives the location within the parent paragraph of the cell's contents.
	StartOffset int

	// EndOFfset gices the location within the parent just after the cell's contents.
	EndOffset int
}

type Line []ParagraphItem

func (l Line) StartOffset() int {
	if len(l) == 0 {
		return -1
	}
	return l[0].StartOffset
}

func (l Line) EndOffset() int {
	if len(l) == 0 {
		return -1
	}
	return l[len(l)-1].EndOffset
}

type Lines []Line

type Layout struct {
	document    *document.Document
	screenWidth int
	paraCache   *lru.Cache[*document.Paragraph, Lines]
}

func (l *Layout) getParagraphLines(p *document.Paragraph) Lines {
	ls, ok := l.paraCache.Get(p)
	if ok {
		return ls
	}

	ls = l.renderParagraphLines(p)

	l.paraCache.Add(p, ls)
	return ls
}

func NewLayout(d *document.Document, screenWidth int) (*Layout, error) {
	paraCache, err := lru.New[*document.Paragraph, Lines](cacheChunkSize)
	if err != nil {
		return nil, err
	}
	return &Layout{
		document:    d,
		screenWidth: screenWidth,
		paraCache:   paraCache,
	}, nil
}

func (l *Layout) ScreenWidth() int {
	return l.screenWidth
}

func (l *Layout) SetScreenWidth(screenWidth int) {
	if screenWidth == l.screenWidth {
		return
	}
	l.paraCache.Purge()
	l.screenWidth = screenWidth
}

func (l *Layout) Document() *document.Document {
	return l.document
}

func (l *Layout) SetDocument(d *document.Document) {
	if d == l.document {
		return
	}
	l.paraCache.Resize((d.ParagraphCount() + cacheChunkSize - 1) & ^(cacheChunkSize - 1))
	l.document = d
}

func (l *Layout) LineIterator(startLineIndex int) *LineIterator {
	return newLineIterator(l, startLineIndex)
}

func (l *Layout) String() string {
	sb := strings.Builder{}

	pitr := l.document.Paragraphs()
	for !pitr.Done() {
		_, p := pitr.Next()
		for _, ln := range l.getParagraphLines(p) {
			for _, item := range ln {
				sb.WriteString(item.Text)
			}
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

func (l *Layout) CellLocationForPoint(p *document.Point) (int, int, error) {
	if p.Document() != l.document {
		return -1, -1, ErrPointIsFromDifferentDocument
	}

	if l.document.ParagraphCount() == 0 {
		return 0, 0, nil
	}

	pitr := p.Document().Paragraphs()
	lineIndex := 0
	targetParaIdx := p.ParagraphIndex()
	targetOffset := p.TextOffset()

	for !pitr.Done() {
		paraIdx, para := pitr.Next()
		lns := l.getParagraphLines(para)

		if paraIdx == targetParaIdx {
			for lnIdx, ln := range lns {
				// TODO: cursor at end of paragraph
				//if targetOffset >= para.TextLength() && ln.EndOffset >= targetOffset {
				//	return len(ln.Cells), lineIndex + lnIdx, nil
				//}

				if ln.StartOffset() > targetOffset || ln.EndOffset() <= targetOffset {
					continue
				}

				x := 0
				for _, item := range ln {
					if item.StartOffset <= targetOffset && item.EndOffset > targetOffset {
						return x + targetOffset - item.StartOffset, lineIndex + lnIdx, nil
					}
					switch item.Type {
					case ParagraphItemTypeBox:
						x += item.CellCount()
					case ParagraphItemTypeGlue:
						x += 1
					}
				}
			}
			return -1, -1, ErrPointNotFound
		}

		lineIndex += len(lns)
	}

	return -1, -1, ErrPointNotFound
}

type LineIterator struct {
	layout        *Layout
	lineIndex     int
	paraIterator  document.ParagraphIterator
	lines         Lines
	paraLineIndex int
}

func newLineIterator(layout *Layout, startLineIndex int) *LineIterator {
	i := &LineIterator{
		layout:       layout,
		paraIterator: layout.document.Paragraphs(),
		lineIndex:    0,
	}

	for !i.paraIterator.Done() {
		_, para := i.paraIterator.Next()
		i.lines = i.layout.getParagraphLines(para)

		if i.lineIndex+len(i.lines) <= startLineIndex {
			i.lineIndex += len(i.lines)
		} else {
			i.paraLineIndex = startLineIndex - i.lineIndex
			i.lineIndex += i.paraLineIndex
			break
		}
	}

	return i
}

func (i *LineIterator) Done() bool {
	if i.lines != nil && i.paraLineIndex < len(i.lines) {
		return false
	}
	return i.paraIterator.Done()
}

func (i *LineIterator) Next() (int, Line) {
	lineIndex := i.lineIndex
	line := i.lines[i.paraLineIndex]

	i.paraLineIndex++
	i.lineIndex++
	for !i.paraIterator.Done() && i.paraLineIndex >= len(i.lines) {
		_, para := i.paraIterator.Next()
		i.lines = i.layout.getParagraphLines(para)
		i.paraLineIndex = 0
	}

	return lineIndex, line
}
