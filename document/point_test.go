package document

import (
	"strings"
	"testing"
)

func docToString(d *Document) string {
	var sb strings.Builder

	for pitr := d.Paragraphs(); !pitr.Done(); {
		_, p := pitr.Next()
		sb.WriteString(p.String())
		if !pitr.Done() {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

func assertDocString(t *testing.T, d *Document, s string) {
	ds := docToString(d)
	if ds != s {
		t.Error("Document not as expected.")
		t.Errorf("Document: %#v", ds)
		t.Errorf("Expected: %#v", s)
	}
}

func TestInsertText(t *testing.T) {
	d := NewDocument()
	d = d.EndPoint().InsertText("Hello").End().InsertText(", world!").Document()
	d = d.StartPoint().Forward().InsertText("yyy").Start().Forward().Forward().InsertText("xxx").Document()
	d = d.EndPoint().InsertText("Goodbye").Document()

	assertDocString(t, d, "Hyyxxxyello, world!\nGoodbye")
}

func TestInsertParaEnd(t *testing.T) {
	d := NewDocument()
	d = d.EndPoint().InsertText("ABC").Document()
	assertDocString(t, d, "ABC")
	d = d.EndPoint().InsertParagraphBreak().End().InsertText("DEF").Document()
	assertDocString(t, d, "ABC\nDEF")
}

func TestInsertPara(t *testing.T) {
	d := NewDocument()
	d = d.StartPoint().InsertText("ABCDEF").Document()
	assertDocString(t, d, "ABCDEF")
	d = d.StartPoint().ForwardN(3).InsertParagraphBreak().Start().InsertText("xy").Document()
	assertDocString(t, d, "ABCxy\nDEF")
	d = d.StartPoint().ForwardN(2).InsertParagraphBreak().End().InsertText("z").Document()
	assertDocString(t, d, "AB\nzCxy\nDEF")
}

func TestForwardToParaEnd(t *testing.T) {
	d := NewDocument()
	d = d.StartPoint().InsertText("ABC").End().InsertParagraphBreak().End().InsertText("DEF").Document()
	d = d.StartPoint().ForwardN(3).InsertText("xyz").Document()
	assertDocString(t, d, "ABCxyz\nDEF")
}

func TestForwardToNextPara(t *testing.T) {
	d := NewDocument()
	d = d.StartPoint().InsertText("ABC").End().InsertParagraphBreak().End().InsertText("DEF").Document()
	d = d.StartPoint().ForwardN(4).InsertText("xyz").Document()
	assertDocString(t, d, "ABC\nxyzDEF")
}

func TestForwardToDocumentEndDoesNotInsertPara(t *testing.T) {
	d := NewDocument()
	d = d.StartPoint().InsertText("ABC").End().InsertParagraphBreak().End().InsertText("DEF").Document()
	d = d.StartPoint().ForwardN(7).InsertText("xyz").End().Forward().InsertText("X").Document()
	assertDocString(t, d, "ABC\nDEFxyzX")
}
