package document

import "github.com/benbjohnson/immutable"

type ParagraphIterator interface {
	Done() bool
	Next() (int, *Paragraph)
}

type Document struct {
	paragraphs *immutable.List[*Paragraph]
	pageWidth  int
}

func NewDocument() *Document {
	return &Document{
		paragraphs: immutable.NewList[*Paragraph](),
		pageWidth:  80,
	}
}

func (d *Document) setParagraph(i int, p *Paragraph) *Document {
	nd := *d
	nd.paragraphs = nd.paragraphs.Set(i, p)
	return &nd
}

func (d *Document) appendParagraph(p *Paragraph) *Document {
	nd := *d
	nd.paragraphs = nd.paragraphs.Append(p)
	return &nd
}

func (d *Document) replaceParagraphs(start int, end int, ps []*Paragraph) *Document {
	lb := immutable.NewListBuilder[*Paragraph]()
	pitr := d.paragraphs.Iterator()

	for !pitr.Done() {
		i, p := pitr.Next()
		if i >= start {
			break
		}
		lb.Append(p)
	}

	for _, p := range ps {
		lb.Append(p)
	}

	if end < d.paragraphs.Len() {
		pitr.Seek(end)
		for !pitr.Done() {
			_, p := pitr.Next()
			lb.Append(p)
		}
	}

	nd := *d
	nd.paragraphs = lb.List()
	return &nd
}

func (d *Document) setParas(ps *immutable.List[*Paragraph]) *Document {
	nd := *d
	nd.paragraphs = ps
	return &nd
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
