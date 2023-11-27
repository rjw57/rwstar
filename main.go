package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rjw57/rwstar/document"
	"github.com/rjw57/rwstar/layout"
)

var rulerStyle = tcell.StyleDefault

/*
func drawRuler(s tcell.Screen, y int, m Margins) {
	w, _ := s.Size()
	pageWidth := int(math.Max(0, float64(w-m.LeftIndent-m.RightIndent)))

	for x := 0; x < w; x++ {
		var c rune

		px := x - m.LeftIndent
		switch {
		case px == 0:
			c = '['
		case px == pageWidth-1:
			c = ']'
		case x == 0:
			c = 'L'
		case x == w-1:
			c = 'R'
		case m.TabStop != 0 && px >= 0 && px < pageWidth && px%m.TabStop == 0:
			c = '|'
		default:
			c = tcell.RuneBullet
		}

		s.SetContent(x, y, c, nil, rulerStyle)
	}
}
*/

func redraw(s tcell.Screen, l *layout.Layout) {
	s.Clear()
	d := l.Document()
	_, h := s.Size()

	y := 0
	pitr := d.Paragraphs()
	for !pitr.Done() && y < h {
		_, p := pitr.Next()
		for _, ln := range l.ParagraphLines(p) {
			for x, cell := range ln {
				s.SetContent(x, y, cell.Mainc, cell.Combc, cell.Style)
			}
			y++
		}
		y++
	}
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	// Set default text style
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)

	// Clear screen
	s.Clear()

	d := document.NewDocument()
	w, _ := s.Size()
	l, err := layout.NewLayout(d, w)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	d = (d.
		StartPoint().
		InsertText("This is an example paragraph. ").End().
		InsertText("This is sentence two of an example paragraph. ").End().
		InsertText("This is sentence three of an example paragraph. ").End().
		InsertText("This is sentence four of an example paragraph. ").End().
		InsertParagraphBreak().End().
		InsertText("This is another example paragraph. ").End().
		InsertText("This is sentence two of another example paragraph. ").End().
		InsertText("This is sentence three of another example paragraph. ").End().
		InsertText("This is sentence four of another example paragraph. ").End().
		InsertParagraphBreak().End().
		InsertText("And another example paragraph.").
		Document())
	l.SetDocument(d)

	redraw(s, l)

	p := d.StartPoint()

	quit := func() {
		s.Fini()
		os.Exit(0)
	}
	for {
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			w, _ = s.Size()
			l.SetScreenWidth(w)
			redraw(s, l)
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC:
				quit()
			case ev.Key() == tcell.KeyEnter:
				p = p.InsertParagraphBreak().End()
				d = p.Document()
				l.SetDocument(d)
				redraw(s, l)

			case ev.Key() == tcell.KeyRune:
				p = p.InsertText(string(ev.Rune())).End()
				d = p.Document()
				l.SetDocument(d)
				redraw(s, l)
			}
		}
	}
}
