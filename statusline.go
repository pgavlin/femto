package femto

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/digitallyserviced/tview"
	"github.com/gdamore/tcell/v2"
)

// Statusline represents the information line at the bottom
// of each view
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type Statusline struct {
	view *View
  line *tview.TextView
}

func NewStatusline(v *View) *Statusline {
  l := tview.NewTextView()
  l.SetDynamicColors(true).SetRegions(true).SetTextAlign(tview.AlignLeft)
  l.SetMaxLines(1)

  sl := &Statusline{
  	view: v,
  	line: l,
  }

  return sl

}

func (sline *Statusline) UpdateScheme(screen tcell.Screen) {

}

// Display draws the statusline to the screen
func (sline *Statusline) Display(screen tcell.Screen) {
	// if messenger.hasPrompt && !GetGlobalOption("infobar").(bool) {
	// 	return
	// }

	// We'll draw the line at the lowest line in the view
	y := sline.view.height-1 + sline.view.y

  bgS := colorscheme.GetColor("line-number")
  _, bg, _ := bgS.Decompose()
  // fmt.Println(bg)
  sline.line.SetBackgroundColor(tcell.GetColor("#21252B"))
  sline.line.SetDontClear(false)

  regions := make([]string, 0)
  makeRegion := func(style, content string) string {
    styleTag := tagColorscheme.GetColorTags(style)
    return fmt.Sprintf(`[%s]%s[-:-:-]`, styleTag, content)

  }

	file := sline.view.Buf.Path
	if sline.view.Buf.Settings["basename"].(bool) {
		file = path.Base(sline.view.Buf.Path)
	}
	// If the buffer is dirty (has been modified) write a little '+'

	if sline.view.Buf.Modified() {
		file += " +"
	}

  regions = append(regions, makeRegion("identifier", fmt.Sprintf(" %s ", file)))

	// Add one to cursor.x and cursor.y because (0,0) is the top left,
	// but users will be used to (1,1) (first line,first column)
	// We use GetVisualX() here because otherwise we get the column number in runes
	// so a '\t' is only 1, when it should be tabSize
	columnNum := strconv.Itoa(sline.view.Cursor.GetVisualX() + 1)
	lineNum := strconv.Itoa(sline.view.Cursor.Y + 1)
  regions = append(regions, makeRegion("line-number", fmt.Sprintf("(%sL,%sC)", lineNum, columnNum)))

	file += " (" + lineNum + "," + columnNum + ")"

	// Add the filetype
	file += " " + sline.view.Buf.FileType()
  regions = append(regions, makeRegion("statusline", fmt.Sprintf(" %s ", sline.view.Buf.FileType())))
  regions = append(regions, makeRegion("statusline", fmt.Sprintf(" %s ", sline.view.Buf.Settings["fileformat"].(string))))

	file += " " + sline.view.Buf.Settings["fileformat"].(string)

	rightText := ""
	// if !sline.view.Buf.Settings["hidehelp"].(bool) {
	// 	if len(kmenuBinding) > 0 {
	// 		if globalSettings["keymenu"].(bool) {
	// 			rightText += kmenuBinding + ": hide bindings"
	// 		} else {
	// 			rightText += kmenuBinding + ": show bindings"
	// 		}
	// 	}
	// 	if len(helpBinding) > 0 {
	// 		if len(kmenuBinding) > 0 {
	// 			rightText += ", "
	// 		}
	// 		if sline.view.Type == vtHelp {
	// 			rightText += helpBinding + ": close help"
	// 		} else {
	// 			rightText += helpBinding + ": open help"
	// 		}
	// 	}
	// 	rightText += " "
	// }
  // sline.line.SetText(strings.Join(regions, makeRegion("divider", "▲ ")))
  sline.line.SetText(strings.Join(regions, makeRegion("statusline", "┇")))
  sline.line.Draw(screen)
  return

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := sline.view.colorscheme["statusline"]; ok {
		statusLineStyle = style
	}
  fg, bg, attr := statusLineStyle.Decompose()
  _,_,_ = fg,bg,attr
  // fmt.Printf("%06x %06x %d", fg.Hex(), bg.Hex(), attr)


	// Maybe there is a unicode filename?
	fileRunes := []rune(file)
	//
	// if sline.view.Type == vtTerm {
	// 	fileRunes = []rune(sline.view.term.title)
	// 	rightText = ""
	// }

	viewX := sline.view.x
	if viewX != 0 {
		screen.SetContent(viewX, y, ' ', nil, statusLineStyle)
		viewX++
	}

  rightStart := 0

	for x := 0; x < sline.view.width; x++ {
		if x < len(fileRunes) {
			screen.SetContent(viewX+x, y, fileRunes[x], nil, statusLineStyle)
		} else if x >= sline.view.width-len(rightText) && x < len(rightText)+sline.view.width-len(rightText) {
			screen.SetContent(viewX+x, y, []rune(rightText)[x-sline.view.width+len(rightText)], nil, statusLineStyle)
		} else {
			screen.SetContent(viewX+x, y, ' ', nil, statusLineStyle)
		}
    if x == sline.view.x + len(fileRunes)+2 {
      rightStart = x + 2
    }
	}

  stLen := tview.TaggedStringWidth(sline.view.statusText)
  if stLen + rightStart <= sline.view.width {
    fg, _,_ := statusLineStyle.Decompose()
    tview.Print(screen, sline.view.statusText, rightStart, y, sline.view.width -rightStart, tview.AlignCenter, fg)
  }
}
