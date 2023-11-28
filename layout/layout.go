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
	ErrorPointIsFromDifferentDocument = errors.New("Point is from different document")
	ErrorPointNotFound                = errors.New("Point not found")
)

type Cell struct {
	Mainc  rune
	Combc  []rune
	Style  tcell.Style
	Offset int
}

type Line struct {
	Cells       []Cell
	StartOffset int
	EndOffset   int
}

type Lines []Line

type Layout struct {
	document    *document.Document
	screenWidth int
	paraCache   *lru.Cache[*document.Paragraph, Lines]
}

func (l *Layout) render(p *document.Paragraph) Lines {
	x := 0
	line := Line{Cells: make([]Cell, 0, 1), StartOffset: 0}
	lines := make(Lines, 0, 0)
	text := p.String()
	for offset, c := range text {
		w := 1
		if x+w > l.screenWidth {
			line.EndOffset = offset
			lines = append(lines, line)
			line = Line{Cells: make([]Cell, 0, 1), StartOffset: offset}
			x = 0
		}
		line.Cells = append(line.Cells, Cell{Mainc: c, Style: tcell.StyleDefault, Offset: offset})
		x += w
	}

	if x > 0 {
		line.EndOffset = len(text)
		lines = append(lines, line)
	}

	// blank line
	lines = append(lines, Line{
		StartOffset: len(text),
		EndOffset:   len(text),
		Cells:       make([]Cell, 0, 0),
	})
	return lines
}

func (l *Layout) getParagraphLines(p *document.Paragraph) Lines {
	ls, ok := l.paraCache.Get(p)
	if ok {
		return ls
	}

	ls = l.render(p)

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
			for _, cell := range ln.Cells {
				sb.WriteRune(cell.Mainc)
			}
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

func (l *Layout) CellLocationForPoint(p *document.Point) (int, int, error) {
	if p.Document() != l.document {
		return -1, -1, ErrorPointIsFromDifferentDocument
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
				if targetOffset >= para.TextLength() && ln.EndOffset >= targetOffset {
					return len(ln.Cells), lineIndex + lnIdx, nil
				}

				if ln.StartOffset > targetOffset || ln.EndOffset <= targetOffset {
					continue
				}

				for x, cell := range ln.Cells {
					if cell.Offset == targetOffset {
						return x, lineIndex + lnIdx, nil
					}
				}
			}
			return -1, -1, ErrorPointNotFound
		}

		lineIndex += len(lns)
	}

	return -1, -1, ErrorPointNotFound
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
