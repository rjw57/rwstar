package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rjw57/rwstar/document"
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

func redraw(s tcell.Screen, d *document.Document) {
	// drawRuler(s, 0, Margins{LeftIndent: 4, RightIndent: 4, TabStop: 8})
}

func docToString(d *document.Document) string {
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

func main() {
	d := (document.NewDocument().
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
	fmt.Println(docToString(d))

	return

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

	redraw(s, d)

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
			redraw(s, d)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				quit()
			}
		}
	}
}
