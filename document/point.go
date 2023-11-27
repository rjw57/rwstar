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

func (p *Point) para() *Paragraph {
	return p.d.paragraphs.Get(p.paraIndex)
}

func (p *Point) reader() *rope.Reader {
	return p.para().text.OffsetReader(p.textOffset)
}

func (p *Point) clone() *Point {
	np := *p
	return &np
}

func (p *Point) withDoc(d *Document) *Point {
	np := *p
	np.d = d
	return &np
}

func (p *Point) Document() *Document {
	return p.d
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
	return p.textOffset >= p.para().text.Length()
}

func (p *Point) DocumentEnd() *Point {
	return p.d.EndPoint()
}

func (p *Point) DocumentStart() *Point {
	return p.d.StartPoint()
}

func (p *Point) Forward() *Point {
	rv := p.clone()

	textLen := rv.para().text.Length()
	if rv.textOffset < textLen {
		r := bufio.NewReader(p.reader())
		_, sz, err := r.ReadRune()
		if err != nil {
			panic(err)
		}
		rv.textOffset += sz
	}

	nPara := rv.d.paragraphs.Len()
	for rv.paraIndex < nPara {
		if rv.textOffset <= rv.para().text.Length() {
			return rv
		}
		rv.textOffset = 0
		rv.paraIndex++
	}

	return p.DocumentEnd()
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
		para := p.para().insertText(p.textOffset, text)
		nd = p.d.setParagraph(p.paraIndex, para)
	}

	start := p.withDoc(nd)
	end := start.clone()
	end.textOffset += len(text)
	return &Range{start: start, end: end}
}

func (p *Point) InsertParagraphBreak() *Range {
	if p.IsDocumentEnd() {
		nd := p.d.appendParagraph(newParagraph(""))
		np := p.withDoc(nd)
		np.textOffset = 0
		return &Range{start: np, end: np}
	}

	lp, rp := p.para().split(p.textOffset)
	nd := p.d.replaceParagraphs(p.paraIndex, p.paraIndex+1, []*Paragraph{lp, rp})

	start := p.withDoc(nd)
	end := start.clone()
	end.paraIndex ++
	end.textOffset = 0

	return &Range{start: start, end: end}
}
