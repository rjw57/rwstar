package document

import "github.com/deadpixi/rope"

type Paragraph struct {
	text rope.Rope
}

func newParagraph(text string) *Paragraph {
	return &Paragraph{text: rope.NewString(text)}
}

func (p *Paragraph) insertText(at int, text string) *Paragraph {
	return &Paragraph{text: p.text.InsertString(at, text)}
}

func (p *Paragraph) split(at int) (*Paragraph, *Paragraph) {
	lt, rt := p.text.Split(at)
	return &Paragraph{text: lt}, &Paragraph{text: rt}
}

func (p *Paragraph) String() string {
	return p.text.String()
}
