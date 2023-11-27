package document

import "github.com/benbjohnson/immutable"

type ParagraphIterator interface {
	Done() bool
	Next() (int, *Paragraph)
}

type Document struct {
	paragraphs *immutable.List[*Paragraph]
}

func NewDocument() *Document {
	return &Document{
		paragraphs: immutable.NewList[*Paragraph](),
	}
}

func (d *Document) setParas(ps *immutable.List[*Paragraph]) *Document {
	return &Document{paragraphs: ps}
}

func (d *Document) StartPoint() *Point {
	return &Point{d: d}
}

func (d *Document) EndPoint() *Point {
	return &Point{d: d, paraIndex: d.paragraphs.Len()}
}

func (d *Document) ParagraphSlice(start int, end int) ParagraphIterator {
	return d.paragraphs.Slice(start, end).Iterator()
}

func (d *Document) Paragraphs() ParagraphIterator {
	return d.paragraphs.Iterator()
}

func (d *Document) GetParagraph(index int) *Paragraph {
	return d.paragraphs.Get(index)
}

func (d *Document) ParagraphCount() int {
	return d.paragraphs.Len()
}
