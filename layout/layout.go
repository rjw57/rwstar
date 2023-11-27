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
	Mainc rune
	Combc []rune
	Style tcell.Style
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
	for _, c := range p.String() {
		w := 1
		if x+w > l.screenWidth {
			lines = append(lines, line)
			line = make(Line, 0, l.screenWidth)
			x = 0
		}
		line = append(line, Cell{Mainc: c, Style: tcell.StyleDefault})
		x += w
	}

	if x > 0 {
		lines = append(lines, line)
	}
	return lines
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

func (l* Layout) Document() *document.Document {
	return l.document
}

func (l *Layout) SetDocument(d *document.Document) {
	if d == l.document {
		return
	}
	l.paraCache.Resize((d.ParagraphCount() + cacheChunkSize - 1) & ^(cacheChunkSize - 1))
	l.document = d
}

func (l *Layout) ParagraphLines(p *document.Paragraph) Lines {
	ls, ok := l.paraCache.Get(p)
	if ok {
		return ls
	}

	ls = l.render(p)

	l.paraCache.Add(p, ls)
	return ls
}

func (l *Layout) String() string {
	sb := strings.Builder{}

	pitr := l.document.Paragraphs()
	for !pitr.Done() {
		_, p := pitr.Next()
		for _, ln := range l.ParagraphLines(p) {
			for _, cell := range ln {
				sb.WriteRune(cell.Mainc)
			}
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}
