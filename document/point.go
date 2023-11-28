package document

import (
	"bufio"

	"github.com/deadpixi/rope"
)

type Point struct {
	d          *Document
	paraIndex  int
	textOffset int
}

func (p *Point) reader() *rope.Reader {
	return p.Paragraph().text.OffsetReader(p.textOffset)
}

func (p *Point) withDoc(d *Document) *Point {
	np := *p
	np.d = d
	return &np
}

func (p *Point) Document() *Document {
	return p.d
}

func (p *Point) Paragraph() *Paragraph {
	if p.paraIndex >= p.d.paragraphs.Len() {
		return nil
	}
	return p.d.paragraphs.Get(p.paraIndex)
}

func (p *Point) ParagraphIndex() int {
	return p.paraIndex
}

func (p *Point) TextOffset() int {
	return p.textOffset
}

func (p *Point) Clone() *Point {
	np := *p
	return &np
}

func (p *Point) IsDocumentEnd() bool {
	return *p == *p.d.EndPoint()
}

func (p *Point) IsDocumentStart() bool {
	return *p == *p.d.StartPoint()
}

func (p *Point) IsParagraphStart() bool {
	return p.textOffset == 0
}

func (p *Point) IsParagraphEnd() bool {
	if p.IsDocumentEnd() {
		return true
	}
	return p.textOffset >= p.Paragraph().TextLength()
}

func (p *Point) DocumentEnd() *Point {
	return p.d.EndPoint()
}

func (p *Point) DocumentStart() *Point {
	return p.d.StartPoint()
}

func (p *Point) Forward() *Point {
	if p.IsDocumentEnd() {
		return p
	}

	rv := p.Clone()

	// Is there still more of the paragraph to go?
	textLen := rv.Paragraph().TextLength()
	if rv.textOffset < textLen {
		r := bufio.NewReader(p.reader())
		_, sz, err := r.ReadRune()
		if err != nil {
			panic(err)
		}
		rv.textOffset += sz
		return rv
	}

	// Do we have another paragraph to move to?
	if rv.paraIndex + 1 >= rv.d.ParagraphCount() {
		// no, move no further
		return rv
	}

	// Move to next paragraph.
	rv.textOffset = 0
	rv.paraIndex++
	nPara := rv.d.paragraphs.Len()
	for rv.paraIndex < nPara && rv.textOffset >= rv.Paragraph().TextLength() {
		rv.paraIndex++
	}

	return rv
}

func (p *Point) ForwardN(n int) *Point {
	for i := 0; i < n; i++ {
		p = p.Forward()
	}
	return p
}

func (p *Point) InsertText(text string) *Range {
	var nd *Document

	if p.IsDocumentEnd() {
		nd = p.d.appendParagraph(newParagraph(text))
	} else {
		para := p.Paragraph().insertText(p.textOffset, text)
		nd = p.d.setParagraph(p.paraIndex, para)
	}

	start := p.withDoc(nd)
	end := start.Clone()
	end.textOffset += len(text)
	return NewRange(start, end)
}

func (p *Point) InsertParagraphBreak() *Range {
	if p.IsDocumentEnd() {
		nd := p.d.appendParagraph(newParagraph(""))
		np := p.withDoc(nd)
		np.textOffset = 0
		return NewRange(np, np)
	}

	lp, rp := p.Paragraph().split(p.textOffset)
	nd := p.d.replaceParagraphs(p.paraIndex, p.paraIndex+1, []*Paragraph{lp, rp})

	start := p.withDoc(nd)
	end := start.Clone()
	end.paraIndex ++
	end.textOffset = 0

	return NewRange(start, end)
}
