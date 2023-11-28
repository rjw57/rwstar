package layout

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/rjw57/rwstar/document"
)

const cacheChunkSizeLog2 = 5
const cacheChunkSize = 1 << cacheChunkSizeLog2

type Cell struct {
	Mainc  rune
	Combc  []rune
	Style  tcell.Style
	Offset int
}

type Line []Cell

type Lines []Line

type Layout struct {
	document    *document.Document
	screenWidth int
	paraCache   *lru.Cache[*document.Paragraph, Lines]
}

func (l *Layout) render(p *document.Paragraph) Lines {
	x := 0
	line := make(Line, 0, l.screenWidth)
	lines := make(Lines, 0, 0)
	for offset, c := range p.String() {
		w := 1
		if x+w > l.screenWidth {
			lines = append(lines, line)
			line = make(Line, 0, l.screenWidth)
			x = 0
		}
		line = append(line, Cell{Mainc: c, Style: tcell.StyleDefault, Offset: offset})
		x += w
	}

	if x > 0 {
		lines = append(lines, line)
	}

	// blank line
	lines = append(lines, make(Line, 0, 0))
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
			for _, cell := range ln {
				sb.WriteRune(cell.Mainc)
			}
			sb.WriteRune('\n')
		}
	}

	return sb.String()
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
	if i.Done() {
		return i.lineIndex, nil
	}

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
